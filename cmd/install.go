package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"

	_ "github.com/lib/pq" // pg driver
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

// These variables are set at build time via ldflags
var eeImage string
var quayImage string
var redisImage string
var postgresImage string

// imageArchivePath is the optional location of the OCI image archive containing required install images
var imageArchivePath string

// executableDir is the optional location of the OCI image archive containing unpacked required install images
var executableDir string

// sshKey is the optional location of the SSH key you would like to use to connect to your host.
var sshKey string

// targetHostname is the hostname of the server you wish to install Quay on
var targetHostname string

// targetUsername is the name of the user on the target host to connect with SSH
var targetUsername string

// initPassword is the password of the initial user.
var initPassword string

// quayHostname is the value to set SERVER_HOSTNAME in the Quay config.yaml
var quayHostname string

// askBecomePass holds whether or not to ask for password during SSH connection
var askBecomePass bool

// additionalArgs are arguments that you would like to append to the end of the ansible-playbook call (used mostly for development)
var additionalArgs string

// installCmd represents the validate command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Quay and its required dependencies.",
	Run: func(cmd *cobra.Command, args []string) {
		install()
	},
}

func init() {

	// Add install command
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&targetHostname, "targetHostname", "H", os.Getenv("HOST"), "The hostname of the target you wish to install Quay to. This defaults to $HOST")
	installCmd.Flags().StringVarP(&targetUsername, "targetUsername", "u", os.Getenv("USER"), "The user on the target host which will be used for SSH. This defaults to $USER")
	installCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/quay_installer", "The path of your ssh identity key. This defaults to ~/.ssh/quay_installer")

	installCmd.Flags().StringVarP(&initPassword, "initPassword", "", "", "The password of the initial user. If not specified, this will be randomly generated.")
	installCmd.Flags().StringVarP(&quayHostname, "quayHostname", "", "", "The value to set SERVER_HOSTNAME in the Quay config.yaml. This defaults to <targetHostname>:8443")

	installCmd.Flags().StringVarP(&imageArchivePath, "image-archive", "i", "", "An archive containing images")
	installCmd.Flags().BoolVarP(&askBecomePass, "askBecomePass", "", false, "Whether or not to ask for password during SSH connection.")
	installCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func install() {

	var err error
	log.Printf("Install has begun")

	log.Debug("Ansible Execution Environment Image: " + eeImage)
	log.Debug("Quay Image: " + quayImage)
	log.Debug("Redis Image: " + redisImage)
	log.Debug("Postgres Image: " + postgresImage)

	// Load execution environment
	err = loadExecutionEnvironment()
	check(err)

	// Check that SSH key is present, and generate if not
	err = loadSSHKeys()
	check(err)

	// Handle Image Archive Defaulting
	var imageArchiveMountFlag string
	if imageArchivePath == "" {
		executableDir, err := os.Executable()
		check(err)
		defaultArchivePath := path.Join(path.Dir(executableDir), "image-archive.tar")
		if pathExists(defaultArchivePath) {
			imageArchivePath = defaultArchivePath
		}
	} else {
		if !pathExists(imageArchivePath) {
			check(errors.New("Could not find image-archive.tar at " + imageArchivePath))
		}
	}

	if imageArchivePath != "" {
		imageArchiveMountFlag = fmt.Sprintf("-v %s:/runner/image-archive.tar", imageArchivePath)
		log.Info("Found image archive at " + imageArchivePath)
		if isLocalInstall() {
			log.Printf("Unpacking image archive from %s", imageArchivePath)
			cmd := exec.Command("tar", "-xvf", imageArchivePath)
			if verbose {
				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
			}
			err = cmd.Run()
			check(err)

			// Load Redis image
			redisArchivePath := path.Join(path.Dir(executableDir), "redis.tar")
			log.Printf("Loading redis image archive from %s", redisArchivePath)
			statement := getImageMetadata("redis", redisImage, redisArchivePath)
			redisImport := exec.Command("/bin/bash", "-c", statement)
			if verbose {
				redisImport.Stderr = os.Stderr
				redisImport.Stdout = os.Stdout
			}
			log.Debug("Importing Redis with command: ", redisImport)
			err = redisImport.Run()
			check(err)

			// Load Postgres image
			postgresArchivePath := path.Join(path.Dir(executableDir), "postgres.tar")
			log.Printf("Loading postgres image archive from %s", postgresArchivePath)
			statement = getImageMetadata("postgres", postgresImage, postgresArchivePath)
			postgresImport := exec.Command("/bin/bash", "-c", statement)
			if verbose {
				postgresImport.Stderr = os.Stderr
				postgresImport.Stdout = os.Stdout
			}
			log.Debug("Importing Postgres with command: ", postgresImport)
			err = postgresImport.Run()
			check(err)

			// Load Quay image
			quayArchivePath := path.Join(path.Dir(executableDir), "quay.tar")
			log.Printf("Loading Quay image archive from %s", quayArchivePath)
			statement = getImageMetadata("quay", quayImage, quayArchivePath)
			quayImport := exec.Command("/bin/bash", "-c", statement)
			if verbose {
				quayImport.Stderr = os.Stderr
				quayImport.Stdout = os.Stdout
			}
			log.Debug("Importing Quay with command: ", quayImport)
			err = quayImport.Run()
			check(err)
		}
		log.Infof("Attempting to set SELinux rules on image archive")
		cmd := exec.Command("chcon", "-Rt", "svirt_sandbox_file_t", imageArchivePath)
		if verbose {
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			log.Warn("Could not set SELinux rule. If your system does not have SELinux enabled, you may ignore this.")
		}
	}

	// Generate password if none provided
	if initPassword == "" {
		initPassword, err = password.Generate(32, 10, 0, false, false)
		check(err)
	}

	// Set quayHostname if not already set
	if quayHostname == "" {
		quayHostname = targetHostname + ":8443"
	}

	// Set askBecomePass flag if true
	var askBecomePassFlag string
	if askBecomePass {
		askBecomePassFlag = "-K"
	}

	// Run playbook
	log.Printf("Running install playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	podmanCmd := fmt.Sprintf(`sudo podman run `+
		`--rm --interactive --tty `+
		`--workdir /runner/project `+
		`--net host `+
		imageArchiveMountFlag+ // optional image archive flag
		` -v %s:/runner/env/ssh_key `+
		`-e RUNNER_OMIT_EVENTS=False `+
		`-e RUNNER_ONLY_FAILED_EVENTS=False `+
		`-e ANSIBLE_HOST_KEY_CHECKING=False `+
		`-e ANSIBLE_CONFIG=/runner/project/ansible.cfg `+
		`--quiet `+
		`--name ansible_runner_instance `+
		eeImage+
		` ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key -e "init_password=%s quay_image=%s redis_image=%s postgres_image=%s quay_hostname=%s local_install=%s" install_mirror_appliance.yml %s %s`,
		sshKey, targetUsername, targetHostname, initPassword, quayImage, redisImage, postgresImage, quayHostname, strconv.FormatBool(isLocalInstall()), askBecomePassFlag, additionalArgs)

	log.Debug("Running command: " + podmanCmd)
	cmd := exec.Command("bash", "-c", podmanCmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	log.Printf("Quay installed successfully")
	log.Printf("Quay is available at %s with credentials (init, %s)", "https://"+quayHostname, initPassword)
}
