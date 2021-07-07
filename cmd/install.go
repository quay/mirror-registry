package cmd

import (
	"fmt"
	"os"
	"os/exec"

	_ "github.com/lib/pq" // pg driver
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// imageArchiveDir is the optional location of the OCI image archive containing required install images
var imageArchiveDir string

// sshKey is the optional location of the SSH key you would like to use to connect to your host.
var sshKey string

// hostname is the hostname of the server you wish to install Quay on
var hostname string

// installCmd represents the validate command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Quay and its required dependencies",
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

}

func install() {

	var err error
	log.Printf("Install has begun")

	log.Infof("Installing ansible-runner")
	err = installAnsibleRunner()
	check(err)

	log.Printf("Attempting to copy ssh file from %s", sshKey)
	cmd := exec.Command("ssh-copy-id", "-i", sshKey, "localhost")
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	check(err)

	cmd = exec.Command("bash", "-c", fmt.Sprintf("podman/tmp/ansible/hacking/env-setup; ansible-playbook -i localhost, --private-key %s /tmp/quay-ansible/p_install-mirror-appliance.yml -kK", sshKey))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	check(err)

	err = uninstallAnsibleRunner()
	check(err)

	// // If image archive is set, load images. Otherwise, pull from dockerhub.
	// if imageArchiveDir == "" { // Flag not set
	// 	// Attempt to autodetect
	// 	executableDir, err := os.Executable()
	// 	if err != nil {
	// 		check(err)
	// 	}
	// 	defaultArchive := path.Join(path.Dir(executableDir), "image-archive.tar")
	// 	if pathExists(defaultArchive) { // Autodetect found archive in same dir as executable
	// 		log.Printf("Autodetected image archive at %s", defaultArchive)
	// 		cmd := exec.Command("podman", "load", "-i", defaultArchive)
	// 		fmt.Print("\033[34m")
	// 		cmd.Stderr = os.Stderr
	// 		cmd.Stdout = os.Stdout
	// 		err = cmd.Run()
	// 		if err != nil {
	// 			check(errors.New(stdErr.String()))
	// 		}
	// 		fmt.Print("\033[0m")
	// 	} else { // No archive provided, pulling images automatically
	// 		log.Printf("Pulling required images")
	// 		for _, s := range services {
	// 			cmd := exec.Command("podman", "pull", s.image)
	// 			fmt.Print("\033[34m")
	// 			cmd.Stderr = os.Stderr
	// 			cmd.Stdout = os.Stdout
	// 			err = cmd.Run()
	// 			if err != nil {
	// 				check(errors.New(stdErr.String()))
	// 			}
	// 			fmt.Print("\033[0m")
	// 		}
	// 	}
	// } else { // Flag was set
	// 	if pathExists(imageArchiveDir) { // Autodetect found archive in same dir as executable
	// 		log.Printf("Using specified image archive at %s", imageArchiveDir)
	// 		cmd := exec.Command("podman", "load", "-i", imageArchiveDir)
	// 		fmt.Print("\033[34m")
	// 		cmd.Stderr = os.Stderr
	// 		cmd.Stdout = os.Stdout
	// 		err = cmd.Run()
	// 		if err != nil {
	// 			check(errors.New(stdErr.String()))
	// 		}
	// 		fmt.Print("\033[0m")
	// 	}
	// }

	// Create podman pod

	log.Printf("Quay installed successfully")
}
