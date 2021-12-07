package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	uninstallCmd.Flags().StringVarP(&sshKey, "ssh-key", "k", os.Getenv("HOME")+"/.ssh/quay_installer", "The path of your ssh identity key. This defaults to ~/.ssh/quay_installer")
	uninstallCmd.Flags().StringVarP(&targetHostname, "targetHostname", "H", "localhost", "The hostname of the target you wish to install Quay to. This defaults to localhost")
	uninstallCmd.Flags().StringVarP(&targetUsername, "targetUsername", "u", os.Getenv("USER"), "The user you wish to ssh into your remote with. This defaults to the current username")
	uninstallCmd.Flags().BoolVarP(&askBecomePass, "askBecomePass", "", false, "Whether or not to ask for sudo password during SSH connection.")
	uninstallCmd.Flags().StringVarP(&additionalArgs, "additionalArgs", "", "", "Additional arguments you would like to append to the ansible-playbook call. Used mostly for development.")

}

func uninstall() {

	var err error
	log.Printf("Uninstall has begun")

	// Load execution environment
	err = loadExecutionEnvironment()
	check(err)

	err = loadSSHKeys()
	check(err)

	// Set askBecomePass flag if true
	var askBecomePassFlag string
	if askBecomePass {
		askBecomePassFlag = "-K"
	}

	log.Printf("Running uninstall playbook. This may take some time. To see playbook output run the installer with -v (verbose) flag.")
	podmanCmd := fmt.Sprintf(`sudo podman run `+
		`--rm --interactive --tty `+
		`--workdir /runner/project `+
		`--net host `+
		` -v %s:/runner/env/ssh_key `+
		`-e RUNNER_OMIT_EVENTS=False `+
		`-e RUNNER_ONLY_FAILED_EVENTS=False `+
		`-e ANSIBLE_HOST_KEY_CHECKING=False `+
		`-e ANSIBLE_CONFIG=/runner/project/ansible.cfg `+
		`--quiet `+
		`--name ansible_runner_instance `+
		eeImage+
		` ansible-playbook -i %s@%s, --private-key /runner/env/ssh_key uninstall_mirror_appliance.yml %s %s`,
		sshKey, targetUsername, strings.Split(targetHostname, ":")[0], askBecomePassFlag, additionalArgs)

	log.Debug("Running command: " + podmanCmd)
	cmd := exec.Command("bash", "-c", podmanCmd)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	check(err)

	log.Printf("Quay uninstalled successfully")
}
