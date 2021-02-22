package cmd

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

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
	log.Printf("Uninstalling Quay")

	var err error
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	// Delete install directory
	installPath := path.Join(os.Getenv("HOME"), "quay-install")
	log.Printf("Searching for Quay install at %s.", installPath)
	if pathExists(installPath) {
		log.Printf("Found Quay install. Deleting directory.")
		check(os.RemoveAll(installPath))
		log.Printf("Deleted Quay install directory.")
	}

	// Delete all services
	for _, s := range services {

		// Delete systemd files
		log.Printf("Searching for %s service at %s.", s.name, s.location)
		if pathExists(s.location) {
			log.Printf("Found %s service. Deleting service file.", s.name)
			check(os.Remove(s.location))
		} else {
			log.Printf("Could not find service file for %s.", s.name)
		}

		// Stop service
		cmd := exec.Command("sudo", "systemctl", "stop", s.name)
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err = cmd.Run()
		if err != nil {
			if strings.Contains(stdErr.String(), "not loaded") {
				log.Printf("Service %s not loaded.", s.name)
			} else {
				check(errors.New(stdErr.String()))
			}
		}

		// Disable service
		cmd = exec.Command("sudo", "systemctl", "disable", s.name)
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err = cmd.Run()
		if err != nil {
			if strings.Contains(stdErr.String(), "does not exist") {
				log.Printf("Service %s not enabled.", s.name)
			} else {
				check(errors.New(stdErr.String()))
			}
		}

		// Reload
		_, err = exec.Command("sudo", "systemctl", "daemon-reload").Output()
		check(err)
		_, err = exec.Command("sudo", "systemctl", "reset-failed").Output()
		check(err)

	}

	log.Printf("Uninstall was successful.")

}
