package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/lib/pq" // pg driver
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

// These variables are set at build time via ldflags
var eeImage string
var pauseImage string
var quayImage string
var redisImage string
var postgresImage string

// imageArchivePath is the optional location of the OCI image archive containing required install images
var imageArchivePath string

// executableDir is the optional location of the OCI image archive containing unpacked required install images
var executableDir string

// sshKey is the optional location of the SSH key you would like to use to connect to your host.
var sshKey string

// sslCert is the path to the SSL certitificate
var sslCert string

// sslKey is the path to the SSL key
var sslKey string

// sslCheckSkip holds whether or not to check the SSL certificate
var sslCheckSkip bool

// targetHostname is the hostname of the server you wish to install Quay on
var targetHostname string

// targetUsername is the name of the user on the target host to connect with SSH
var targetUsername string

// initUser is the initial username.
var initUser string

// initPassword is the password of the initial user.
var initPassword string

// quayHostname is the value to set SERVER_HOSTNAME in the Quay config.yaml
var quayHostname string

// askBecomePass holds whether or not to ask for password during SSH connection
var askBecomePass bool

// quayRoot is the directory where all the data are stored
var quayRoot string

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

	installCmd.Flags().StringVarP(&targetHostname, "targetHostname", "H", getFQDN(), "The hostname of the target you wish to install Quay to. This defaults to $HOST")
	installCmd.Flags().StringVarP(&targetUsername, "targetUsername", "u", os.Getenv("USER"), "The user on the target host which will be used for SSH. This defaults to $USER")
	installCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/quay_installer", "The path of your ssh identity key. This defaults to ~/.ssh/quay_installer")

	installCmd.Flags().StringVarP(&sslCert, "sslCert", "", "", "The path to the SSL certificate Quay should use")
	installCmd.Flags().StringVarP(&sslKey, "sslKey", "", "", "The path to the SSL key Quay should use")
	installCmd.Flags().BoolVarP(&sslCheckSkip, "sslCheckSkip", "", false, "Whether or not to check the certificate hostname against the SERVER_HOSTNAME in config.yaml.")

	installCmd.Flags().StringVarP(&initUser, "initUser", "", "init", "The password of the initial user. This defaults to init.")
	installCmd.Flags().StringVarP(&initPassword, "initPassword", "", "", "The password of the initial user. If not specified, this will be randomly generated.")
	installCmd.Flags().StringVarP(&quayHostname, "quayHostname", "", "", "The value to set SERVER_HOSTNAME in the Quay config.yaml. This defaults to <targetHostname>:8443")

	installCmd.Flags().StringVarP(&imageArchivePath, "image-archive", "i", "", "An archive containing images")
	installCmd.Flags().BoolVarP(&askBecomePass, "askBecomePass", "", false, "Whether or not to ask for sudo password during SSH connection.")
	installCmd.Flags().StringVarP(&quayRoot, "quayRoot", "r", "/etc/quay-install", "The folder where quay persistent data are saved. This defaults to /etc/quay-install")
	installCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func install() {

	var err error
	log.Printf("Install has begun")

	log.Debug("Ansible Execution Environment Image: " + eeImage)
	log.Debug("Pause Image: " + pauseImage)
	log.Debug("Quay Image: " + quayImage)
	log.Debug("Redis Image: " + redisImage)
	log.Debug("Postgres Image: " + postgresImage)

	// Load execution environment
	err = loadExecutionEnvironment()
	check(err)

	// Set quayHostname if not already set
	if quayHostname == "" {
		quayHostname = targetHostname + ":8443"
	}

	// Load the SSL certificate and the key
	err = loadCerts(sslCert, sslKey, strings.Split(quayHostname, ":")[0], sslCheckSkip)
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

			// Load Pause image
			pauseArchivePath := path.Join(path.Dir(executableDir), "pause.tar")
			log.Printf("Loading pause image archive from %s", pauseArchivePath)
			statement := getImageMetadata("pause", pauseImage, pauseArchivePath)
			pauseImport := exec.Command("/bin/bash", "-c", statement)
			if verbose {
				pauseImport.Stderr = os.Stderr
				pauseImport.Stdout = os.Stdout
			}
			log.Debug("Importing Pause with command: ", pauseImport)
			err = pauseImport.Run()
			check(err)

			// Load Redis image
			redisArchivePath := path.Join(path.Dir(executableDir), "redis.tar")
			log.Printf("Loading redis image archive from %s", redisArchivePath)
			statement = getImageMetadata("redis", redisImage, redisArchivePath)
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

	// Add port if not present
	if !strings.Contains(quayHostname, ":") {
		quayHostname = quayHostname + ":8443"
	}

	// Set askBecomePass flag if true
	var askBecomePassFlag string
	if askBecomePass {
		askBecomePassFlag = "-K"
	}

	// Set the SSL flag if cert and key are defined
	var sslCertKeyFlag string
	if sslCert != "" && sslKey != "" {
		sslCertAbs, err := filepath.Abs(sslCert)
		if err != nil {
			check(errors.New("Unable to get absolute path of " + sslCert))
		}
		sslKeyAbs, err := filepath.Abs(sslKey)
		if err != nil {
			check(errors.New("Unable to get absolute path of " + sslKey))
		}
		sslCertKeyFlag = fmt.Sprintf(" -v %s:/runner/certs/quay.cert:Z -v %s:/runner/certs/quay.key:Z", sslCertAbs, sslKeyAbs)
	}

	// Run playbook
	log.Printf("Running install playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	podmanCmd := fmt.Sprintf(`sudo podman run `+
		`--rm --interactive --tty `+
		`--workdir /runner/project `+
		`--net host `+
		imageArchiveMountFlag+ // optional image archive flag
		sslCertKeyFlag+ // optional ssl cert/key flag
		` -v %s:/runner/env/ssh_key `+
		`-e RUNNER_OMIT_EVENTS=False `+
		`-e RUNNER_ONLY_FAILED_EVENTS=False `+
		`-e ANSIBLE_HOST_KEY_CHECKING=False `+
		`-e ANSIBLE_CONFIG=/runner/project/ansible.cfg `+
		`--quiet `+
		`--name ansible_runner_instance `+
		fmt.Sprintf("%s ", eeImage)+
		`ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key -e "init_user=%s init_password=%s quay_image=%s redis_image=%s postgres_image=%s pause_image=%s quay_hostname=%s local_install=%s quay_root=%s" install_mirror_appliance.yml %s %s`,
		sshKey, targetUsername, targetHostname, initUser, initPassword, quayImage, redisImage, postgresImage, pauseImage, quayHostname, strconv.FormatBool(isLocalInstall()), quayRoot, askBecomePassFlag, additionalArgs)

	log.Debug("Running command: " + podmanCmd)
	cmd := exec.Command("bash", "-c", podmanCmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	log.Printf("Quay installed successfully, permanent data is stored in %s", quayRoot)
	log.Printf("Quay is available at %s with credentials (%s, %s)", "https://"+quayHostname, initUser, initPassword)
}
