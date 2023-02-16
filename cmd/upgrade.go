package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	_ "github.com/lib/pq" // pg driver
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade all mirror registry images.",
	Run: func(cmd *cobra.Command, args []string) {
		upgrade()
	},
}

func init() {

	// Add upgrade command
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().StringVarP(&targetHostname, "targetHostname", "H", getFQDN(), "The hostname of the target you wish to install Quay to. This defaults to $HOST")
	upgradeCmd.Flags().StringVarP(&targetUsername, "targetUsername", "u", os.Getenv("USER"), "The user on the target host which will be used for SSH. This defaults to $USER")
	upgradeCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/quay_installer", "The path of your ssh identity key. This defaults to ~/.ssh/quay_installer")

	upgradeCmd.Flags().StringVarP(&quayHostname, "quayHostname", "", "", "The value to set SERVER_HOSTNAME in the Quay config.yaml. This defaults to <targetHostname>:8443")

	upgradeCmd.Flags().StringVarP(&imageArchivePath, "image-archive", "i", "", "An archive containing images")
	upgradeCmd.Flags().BoolVarP(&askBecomePass, "askBecomePass", "", false, "Whether or not to ask for sudo password during SSH connection.")
	upgradeCmd.Flags().StringVarP(&quayRoot, "quayRoot", "r", "~/quay-install", "The folder where quay persistent data are saved. This defaults to ~/quay-install")
	upgradeCmd.Flags().StringVarP(&quayStorage, "quayStorage", "", "quay-storage", "The folder where quay persistent storage data is saved. This defaults to a Podman named volume 'quay-storage'. Root is required to uninstall.")
	upgradeCmd.Flags().StringVarP(&pgStorage, "pgStorage", "", "pg-storage", "The folder where postgres persistent storage data is saved. This defaults to a Podman named volume 'pg-storage'. Root is required to uninstall.")
	upgradeCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func upgrade() {

	var err error
	log.Printf("Upgrade has begun")

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

	// Run playbook
	log.Printf("Running upgrade playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	quayVersion := strings.Split(quayImage, ":")[1]
	podmanCmd := fmt.Sprintf(`podman run `+
		`--rm --interactive --tty `+
		`--workdir /runner/project `+
		`--net host `+
		imageArchiveMountFlag+ // optional image archive flag
		` -v %s:/runner/env/ssh_key `+
		`-e RUNNER_OMIT_EVENTS=False `+
		`-e RUNNER_ONLY_FAILED_EVENTS=False `+
		`-e ANSIBLE_HOST_KEY_CHECKING=False `+
		`-e ANSIBLE_CONFIG=/runner/project/ansible.cfg `+
		fmt.Sprintf("-e ANSIBLE_NOCOLOR=%t ", noColor)+
		`--quiet `+
		`--name ansible_runner_instance `+
		fmt.Sprintf("%s ", eeImage)+
		`ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key -e "quay_image=%s quay_version=%s redis_image=%s postgres_image=%s pause_image=%s quay_hostname=%s local_install=%s quay_root=%s quay_storage=%s pg_storage=%s" upgrade_mirror_appliance.yml %s %s`,
		sshKey, targetUsername, targetHostname, quayImage, quayVersion, redisImage, postgresImage, pauseImage, quayHostname, strconv.FormatBool(isLocalInstall()), quayRoot, quayStorage, pgStorage, askBecomePassFlag, additionalArgs)

	log.Debug("Running command: " + podmanCmd)
	cmd := exec.Command("bash", "-c", podmanCmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	log.Printf("Quay upgraded successfully")
}
