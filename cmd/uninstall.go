package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall will remove all Quay dependencies.",
	Run: func(cmd *cobra.Command, args []string) {
		uninstall()
	},
}

func init() {

	// Add install command
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/id_rsa", "The path of your ssh identity key. This defaults to ~/.ssh/id_rsa")
	uninstallCmd.Flags().StringVarP(&targetHostname, "targetHostname", "H", "localhost", "The hostname of the target you wish to install Quay to. This defaults to localhost")
	uninstallCmd.Flags().StringVarP(&targetUsername, "targetUsername", "u", os.Getenv("USER"), "The user you wish to ssh into your remote with. This defaults to the current username")
	uninstallCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "-K", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func uninstall() {

	var err error
	log.Printf("Uninstall has begun")

	executableDir, err := os.Executable()
	if err != nil {
		check(err)
	}

	log.Printf("Loading execution environment from execution-environment.tar")
	cmd := exec.Command("sudo", "podman", "load", "-i", path.Join(path.Dir(executableDir), "execution-environment.tar"))
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	err = cmd.Run()
	check(err)

	log.Printf("Running uninstall playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	cmd = exec.Command("bash", "-c", fmt.Sprintf(`sudo podman run --rm --tty --interactive --workdir /runner/project --net host -v %s:/runner/env/ssh_key  --quiet --name ansible_runner_instance quay.io/quay/openshift-mirror-registry-ee ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key uninstall_mirror_appliance.yml %s`, sshKey, targetUsername, targetHostname, additionalArgs))
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	cleanup()

	log.Printf("Quay uninstalled successfully")
}
