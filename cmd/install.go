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

// These variables are set during compilation time
var quayImage string
var redisImage string
var postgresImage string

// imageArchivePath is the optional location of the OCI image archive containing required install images
var imageArchivePath string

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

// askBecomePass holds whether or not to ask for sudo password during SSH connection
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
	installCmd.Flags().BoolVarP(&askBecomePass, "askBecomePass", "", false, "Whether or not to ask for sudo password during SSH connection.")
	installCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func install() {

	var err error
	log.Printf("Install has begun")

	log.Debug("Quay Image: " + quayImage)
	log.Debug("Redis Image: " + redisImage)
	log.Debug("Postgres Image: " + postgresImage)

	// Detect if installation is local
	var localInstall bool
	if targetHostname == "localhost" && targetUsername == os.Getenv("USER") {
		log.Infof("Detected an installation to localhost")
		localInstall = true
	}

	// Check that executable environment is present
	executableDir, err := os.Executable()
	check(err)
	executionEnvironmentPath := path.Join(path.Dir(executableDir), "execution-environment.tar")
	if !pathExists(executionEnvironmentPath) {
		check(errors.New("Could not find execution-environment.tar at " + executionEnvironmentPath))
	}
	log.Info("Found execution environment at " + executionEnvironmentPath)

	// Check that SSH key is present, and generate if not
	if sshKey == os.Getenv("HOME")+"/.ssh/quay_installer" && localInstall {
		if pathExists(sshKey) {
			log.Info("Found SSH key at " + sshKey)
		} else {
			log.Info("Did not find SSH key in default location. Attempting to set up SSH keys.")
			err = setupLocalSSH(targetHostname, targetUsername)
			check(err)
			log.Info("Successfully set up SSH keys")
		}
	} else {
		if !pathExists(sshKey) {
			check(errors.New("Could not find ssh key at " + sshKey))
		} else {
			log.Info("Found SSH key at " + sshKey)
		}
	}

	// Handle Image Archive Loading/Defaulting
	var imageArchiveMountFlag string
	if imageArchivePath == "" {
		defaultArchivePath := path.Join(path.Dir(executableDir), "image-archive.tar")
		if pathExists(defaultArchivePath) {
			imageArchiveMountFlag = fmt.Sprintf("-v %s:/runner/image-archive.tar", defaultArchivePath)
			log.Info("Found image archive at " + defaultArchivePath)
			if localInstall {
				log.Printf("Loading image archive from %s", defaultArchivePath)
				cmd := exec.Command("sudo", "podman", "load", "-i", defaultArchivePath)
				if verbose {
					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout
				}
				err = cmd.Run()
				check(err)
			}
		}
	} else { // Flag was set
		if pathExists(imageArchivePath) {
			imageArchiveMountFlag = fmt.Sprintf("-v %s:/runner/image-archive.tar", imageArchivePath)
			log.Info("Found image archive at " + imageArchivePath)
			if localInstall {
				log.Printf("Loading image archive from %s", imageArchivePath)
				cmd := exec.Command("sudo", "podman", "load", "-i", imageArchivePath)
				if verbose {
					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout
				}
				err = cmd.Run()
				check(err)
			}
		} else {
			check(errors.New("Could not find image-archive.tar at " + imageArchivePath))
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

	// Load execution environment into podman
	log.Printf("Loading execution environment from execution-environment.tar")
	cmd := exec.Command("sudo", "podman", "load", "-i", executionEnvironmentPath)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	err = cmd.Run()
	check(err)

	// FIXME - find a better way to collect logs
	// Create log file to collect logs
	// logFile, err := ioutil.TempFile("", "ansible-output")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Debug("Writing ansible playbook logs to " + logFile.Name())
	// defer os.Remove(logFile.Name())

	// go watchFileAndRun(logFile.Name())

	// Run playbook
	log.Printf("Running install playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	podmanCmd := fmt.Sprintf(`sudo podman run `+
		`--rm --interactive --tty `+
		`--workdir /runner/project `+
		`--net host `+
		imageArchiveMountFlag+ // optional image archive flag
		` -v %s:/runner/env/ssh_key `+
		// `-v %s:/var/log/ansible/hosts/`+targetUsername+`@`+targetHostname+` `+
		`-e RUNNER_OMIT_EVENTS=False `+
		`-e RUNNER_ONLY_FAILED_EVENTS=False `+
		`-e ANSIBLE_HOST_KEY_CHECKING=False `+
		`-e ANSIBLE_CONFIG=/runner/project/ansible.cfg `+
		// `-e ANSIBLE_STDOUT_CALLBACK=log_plays `+
		`--quiet `+
		`--name ansible_runner_instance `+
		`quay.io/quay/openshift-mirror-registry-ee `+
		// FIXME - Put extra variables into a temp file and then mount into /runner/env?
		`ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key -e "init_password=%s quay_image=%s redis_image=%s postgres_image=%s quay_hostname=%s local_install=%s" install_mirror_appliance.yml %s %s`,
		sshKey, targetUsername, targetHostname, initPassword, quayImage, redisImage, postgresImage, quayHostname, strconv.FormatBool(localInstall), askBecomePassFlag, additionalArgs)

	log.Debug("Running command: " + podmanCmd)
	cmd = exec.Command("bash", "-c", podmanCmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	cleanup()
	log.Printf("Quay installed successfully")
	log.Printf("Quay is available at %s with credentials (init, %s)", "https://"+quayHostname, initPassword)
}
