package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	_ "github.com/lib/pq" // pg driver
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// imageArchiveDir is the optional location of the OCI image archive containing required install images
var imageArchiveDir string

// sshKey is the optional location of the SSH key you would like to use to connect to your host.
var sshKey string

// hostname is the hostname of the server you wish to install Quay on
var hostname string

// username is the name of the user which you wish to SSH into the remote with.
var username string

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

	installCmd.Flags().StringVarP(&imageArchiveDir, "image-archive", "i", "", "An archive containing images")
	installCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/id_rsa", "The path of your ssh identity key. This defaults to ~/.ssh/id_rsa")
	installCmd.Flags().StringVarP(&hostname, "hostname", "H", "localhost", "The hostname you wish to install Quay to. This defaults to localhost")
	installCmd.Flags().StringVarP(&username, "username", "u", os.Getenv("USER"), "The user you wish to ssh into your remote with. This defaults to the current username")
	installCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "-K", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func install() {

	var err error
	log.Printf("Install has begun")

	// Check that all files are present
	executableDir, err := os.Executable()
	check(err)
	executionEnvironmentPath := path.Join(path.Dir(executableDir), "execution-environment.tar")
	if !pathExists(executionEnvironmentPath) {
		check(errors.New("Could not find execution-environment.tar at " + executionEnvironmentPath))
	}
	if !pathExists(sshKey) {
		check(errors.New("Could not find ssh key at " + sshKey))
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

	// Generate password
	generatedPassword, err := password.Generate(32, 10, 0, false, false)
	check(err)

	// Run playbook
	log.Printf("Running install playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	cmd = exec.Command("bash", "-c", fmt.Sprintf(`sudo podman run --rm --tty --interactive --workdir /runner/project --net host -v %s:/runner/env/ssh_key  --quiet --name ansible_runner_instance quay.io/quay/openshift-mirror-registry-ee ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key -e "init_password=%s" install_mirror_appliance.yml %s`, sshKey, username, hostname, generatedPassword, additionalArgs))
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	cleanup()
	log.Printf("Quay installed successfully")
	log.Printf("Quay is available at %s with credentials (init, %s)", "https://"+hostname, generatedPassword)
}
