package cmd

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall will remove all Quay dependencies from your host machine",
	Run: func(cmd *cobra.Command, args []string) {
		uninstall()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func uninstall() {
	log.Printf("Uninstall has begun")

	log.Printf("Installing ansible-runner")
	err := installAnsibleRunner()
	check(err)

	log.Printf("Attempting to copy ssh file from %s", sshKey)
	cmd := exec.Command("ssh-copy-id", "-i", sshKey, "localhost")
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	check(err)

	cmd = exec.Command("bash", "-c", fmt.Sprintf("source /tmp/ansible/hacking/env-setup; ansible-playbook -i localhost, --private-key %s /tmp/quay-ansible/p_uninstall-mirror-appliance.yml -kK", sshKey))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	check(err)

	err = uninstallAnsibleRunner()
	check(err)

}
