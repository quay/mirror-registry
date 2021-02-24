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
	installPath := path.Join("/root", "quay-install") // FIXME - find a better way to set this path.
	log.Printf("Searching for Quay install at %s.", installPath)
	if pathExists(installPath) {
		log.Printf("Found Quay install. Deleting directory.")
		check(os.RemoveAll(installPath))
		log.Printf("Deleted Quay install directory.")
	}

	// Reload daemon
	_, err = exec.Command("sudo", "systemctl", "daemon-reexec").Output()
	check(err)

	// Delete all services
	for _, s := range services {

		// Stop service
		log.Printf("Stopping service %s.", s.name)
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
		} else {
			log.Printf("Stopped service %s.", s.name)
		}

		// Disable service
		log.Printf("Disabling service %s.", s.name)
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
		} else {
			log.Printf("Disabled %s disabled.", s.name)
		}

		// Delete systemd files
		log.Printf("Searching for %s service at %s.", s.name, s.location)
		if pathExists(s.location) {
			log.Printf("Found %s service. Deleting service file.", s.name)
			check(os.Remove(s.location))
		} else {
			log.Printf("Could not find service file for %s.", s.name)
		}

		// Reload
		_, err = exec.Command("sudo", "systemctl", "daemon-reload").Output()
		check(err)
		_, err = exec.Command("sudo", "systemctl", "reset-failed").Output()
		check(err)

	}

	log.Printf("Uninstall was successful.")

}
