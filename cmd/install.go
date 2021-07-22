package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

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

	installCmd.Flags().StringVarP(&targetHostname, "targetHostname", "H", "localhost", "The hostname of the target you wish to install Quay to. This defaults to localhost")
	installCmd.Flags().StringVarP(&targetUsername, "targetUsername", "u", os.Getenv("USER"), "The user on the target host which will be used for SSH. This defaults to the current username")
	installCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/id_rsa", "The path of your ssh identity key. This defaults to ~/.ssh/id_rsa")

	installCmd.Flags().StringVarP(&initPassword, "initPassword", "c", "", "The password of the initial user. If not specified, this will be randomly generated.")

	installCmd.Flags().StringVarP(&imageArchivePath, "image-archive", "i", "", "An archive containing images")
	installCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "-K", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func install() {

	var err error
	log.Printf("Install has begun")

	log.Debug("Quay Image: " + quayImage)
	log.Debug("Redis Image: " + redisImage)
	log.Debug("Postgres Image: " + postgresImage)

	// Check that all files are present
	executableDir, err := os.Executable()
	check(err)
	executionEnvironmentPath := path.Join(path.Dir(executableDir), "execution-environment.tar")
	if !pathExists(executionEnvironmentPath) {
		check(errors.New("Could not find execution-environment.tar at " + executionEnvironmentPath))
	}
	log.Info("Found execution environment at " + executionEnvironmentPath)
	if !pathExists(sshKey) {
		check(errors.New("Could not find ssh key at " + sshKey))
	}
	log.Info("Found SSH key at " + sshKey)

	// Handle Image Archive Loading/Defaulting
	var imageArchiveMountFlag string
	if imageArchivePath == "" {
		defaultArchivePath := path.Join(path.Dir(executableDir), "image-archive.tar")
		if pathExists(defaultArchivePath) {
			imageArchiveMountFlag = fmt.Sprintf("-v %s:/runner/image-archive.tar", defaultArchivePath)
			log.Info("Found image archive at " + defaultArchivePath)
		}
	} else { // Flag was set
		if pathExists(imageArchivePath) {
			imageArchiveMountFlag = fmt.Sprintf("-v %s:/runner/image-archive.tar", imageArchivePath)
			log.Info("Found image archive at " + imageArchivePath)
		} else {
			check(errors.New("Could not find image-archive.tar at " + imageArchivePath))
		}
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

	// Generate password if none provided
	if initPassword == "" {
		initPassword, err = password.Generate(32, 10, 0, false, false)
		check(err)
	}

	// Create log file to collect logs
	logFile, err := ioutil.TempFile("", "ansible-output")
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Writing ansible playbook logs to " + logFile.Name())
	defer os.Remove(logFile.Name())

	// go watchFileAndRun(logFile.Name())

	// Run playbook
	log.Printf("Running install playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	podmanCmd := fmt.Sprintf(`sudo podman run `+
		`--rm --interactive --tty `+
		`--workdir /runner/project `+
		`--net host `+
		imageArchiveMountFlag+ // optional image archive flag
		` -v %s:/runner/env/ssh_key `+
		`-v %s:/var/log/ansible/hosts/`+targetUsername+`@`+targetHostname+` `+
		`-e ANSIBLE_CACHE_PLUGIN=jsonfile `+
		`-e ANSIBLE_CACHE_PLUGIN_CONNECTION=/runner/artifacts/instance/fact_cache `+
		`-e AWX_ISOLATED_DATA_DIR=/runner/artifacts/instance `+
		`-e RUNNER_OMIT_EVENTS=False `+
		`-e RUNNER_ONLY_FAILED_EVENTS=False `+
		`-e ANSIBLE_HOST_KEY_CHECKING=False `+
		`-e LAUNCHED_BY_RUNNER=1 `+
		// `-e ANSIBLE_STDOUT_CALLBACK=log_plays `+
		`-e ANSIBLE_RETRY_FILES_ENABLED=False `+
		`--quiet `+
		`--name ansible_runner_instance `+
		`quay.io/quay/openshift-mirror-registry-ee `+
		`ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key -e "init_password=%s quay_image=%s redis_image=%s postgres_image=%s" install_mirror_appliance.yml %s`,
		sshKey, logFile.Name(), targetUsername, strings.Split(targetHostname, ":")[0], initPassword, quayImage, redisImage, postgresImage, additionalArgs)

	log.Debug("Running command: " + podmanCmd)
	cmd = exec.Command("bash", "-c", podmanCmd)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	cleanup()
	log.Printf("Quay installed successfully")
	log.Printf("Quay is available at %s with credentials (init, %s)", "https://"+targetHostname, initPassword)
}
