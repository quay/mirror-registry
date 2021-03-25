package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

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

	// Delete podman pod
	log.Printf("Deleting pod for Quay containers.")
	cmd := exec.Command("podman", "pod", "rm", "--force", "quay-pod")
	fmt.Print("\033[34m")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		check(errors.New(stdErr.String()))
	}
	log.Printf("Deleted pod for Quay containers.")

	// Reload daemon
	_, err = exec.Command("systemctl", "daemon-reexec").Output()
	check(err)

	// Delete all services
	for _, s := range services {

		// Stop service
		log.Printf("Stopping service %s.", s.name)
		cmd := exec.Command("systemctl", "stop", s.name)
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err = cmd.Run()
		if err != nil {
			if strings.Contains(stdErr.String(), "not loaded") {
				log.Warningf("Service %s not loaded.", s.name)
			} else {
				check(errors.New(stdErr.String()))
			}
		} else {
			log.Printf("Stopped service %s.", s.name)
		}

		// Disable service
		log.Printf("Disabling service %s.", s.name)
		cmd = exec.Command("systemctl", "disable", s.name)
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err = cmd.Run()
		if err != nil {
			if strings.Contains(stdErr.String(), "does not exist") {
				log.Warningf("Service %s not enabled.", s.name)
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
			log.Warningf("Could not find service file for %s.", s.name)
		}

		// Reload
		_, err = exec.Command("systemctl", "daemon-reload").Output()
		check(err)
		_, err = exec.Command("systemctl", "reset-failed").Output()
		check(err)

	}

	log.Printf("Uninstall was successful.")

}
