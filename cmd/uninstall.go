package cmd

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

// editorCmd represents the validate command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Runs a browser-based editor for your config.yaml",

	Run: func(cmd *cobra.Command, args []string) {
		uninstall()
	},
}

func init() {
	// Add editor command
	rootCmd.AddCommand(uninstallCmd)

	// // Add --config-dir flag
	// editorCmd.Flags().StringVarP(&configDir, "config-dir", "c", "", "The directory containing your config files")

	// // Add --password flag
	// editorCmd.Flags().StringVarP(&editorPassword, "password", "p", "", "The password to enter the editor")
	// editorCmd.MarkFlagRequired("password")

	// // Add --operator-endpoint flag
	// editorCmd.Flags().StringVarP(&operatorEndpoint, "operator-endpoint", "e", "", "The endpoint to commit a validated config bundle to")

	// // Add --readonly-fieldgroups flag
	// editorCmd.Flags().StringVarP(&readonlyFieldGroups, "readonly-fieldgroups", "r", "", "Comma-separated list of fieldgroups that should be treated as read-only")
}

func uninstall() {

	log.Printf("Uninstalling")
	// Set permissions

	_, err := exec.Command("systemctl", "stop", "quay-postgresql").Output()
	check(err)
	_, err = exec.Command("systemctl", "disable", "quay-postgresql").Output()
	check(err)

	if _, err := os.Stat(postgresqlSystemdLocation); err == nil {
		err = os.Remove(postgresqlSystemdLocation)
		check(err)
	}

	_, err = exec.Command("systemctl", "daemon-reload").Output()
	check(err)
	_, err = exec.Command("systemctl", "reset-failed").Output()
	check(err)

	installPath := path.Join(os.Getenv("HOME"), "quay-install")
	err = os.RemoveAll(installPath)
	check(err)

}
