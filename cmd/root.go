package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Create logger
var log = &logrus.Logger{
	Out:   os.Stdout,
	Level: logrus.InfoLevel,
}

func loadExecutionEnvironment() error {

	// Ensure execution environment is present
	executableDir, err := os.Executable()
	if err != nil {
		return err
	}
	executionEnvironmentPath := path.Join(path.Dir(executableDir), "execution-environment.tar")
	if !pathExists(executionEnvironmentPath) {
		return errors.New("Could not find execution-environment.tar at " + executionEnvironmentPath)
	}
	log.Info("Found execution environment at " + executionEnvironmentPath)

	// Load execution environment into podman
	log.Printf("Loading execution environment from execution-environment.tar")
	cmd := exec.Command("sudo", "podman", "load", "-i", executionEnvironmentPath)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func isLocalInstall() bool {
	if targetHostname == "localhost" && targetUsername == os.Getenv("USER") {
		log.Infof("Detected an installation to localhost")
		return true
	}
	return false
}

func loadSSHKeys() error {
	if sshKey == os.Getenv("HOME")+"/.ssh/quay_installer" && isLocalInstall() {
		if pathExists(sshKey) {
			log.Info("Found SSH key at " + sshKey)
		} else {
			log.Info("Did not find SSH key in default location. Attempting to set up SSH keys.")
			if err := setupLocalSSH(); err != nil {
				return err
			}
			log.Info("Successfully set up SSH keys")
		}
	} else {
		if !pathExists(sshKey) {
			return errors.New("Could not find ssh key at " + sshKey)
		} else {
			log.Info("Found SSH key at " + sshKey)
		}
	}
	setSELinux(sshKey)

    return nil
}

func setupLocalSSH() error {

	log.Infof("Generating SSH Key")
	cmd := exec.Command("bash", "-c", "ssh-keygen -b 2048 -t rsa -N '' -f ~/.ssh/quay_installer")
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return err
	}
	log.Infof("Generated SSH Key at " + os.Getenv("HOME") + "/.ssh/quay_installer")

	keyFile, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/quay_installer.pub")
	if err != nil {
		return err
	}

	log.Infof("Adding key to ~/.ssh/authorized_keys")
	cmd = exec.Command("bash", "-c", "/bin/echo \""+string(keyFile)+"\" >> ~/.ssh/authorized_keys")
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func loadCerts() error {
	if pathExists(sslCert) {
		log.Info("Found SSL certificate at " + sslCert)
	} else {
		log.Info("Did not found SSL certificate at " + sslCert)
		return errors.New("Could not find SSL certificate at " + sslCert)
	}
	setSELinux(sslCert)

	if pathExists(sslKey) {
		log.Info("Found SSL key at " + sslKey)
	} else {
		log.Info("Did not found SSL key at " + sslKey)
		return errors.New("Could not find SSL key at " + sslKey)
	}
	setSELinux(sslKey)

	return nil
}

func setSELinux(path string) {
	log.Infof("Attempting to set SELinux rules on " + path)
	cmd := exec.Command("chcon", "-Rt", "svirt_sandbox_file_t", path)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		log.Warn("Could not set SELinux rule. If your system does not have SELinux enabled, you may ignore this.")
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return true
	}
	return false
}

func check(err error) {
	if err != nil {
		log.Errorf("An error occurred: %s", err.Error())
		os.Exit(1)
	}
}

// verbose is the optional command that will display INFO logs
var verbose bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose logs")
}

var (
	rootCmd = &cobra.Command{
		Use: "openshift-mirror-registry",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	fmt.Println(`
   __   __
  /  \ /  \     ______   _    _     __   __   __
 / /\ / /\ \   /  __  \ | |  | |   /  \  \ \ / /
/ /  / /  \ \  | |  | | | |  | |  / /\ \  \   /
\ \  \ \  / /  | |__| | | |__| | / ____ \  | |
 \ \/ \ \/ /   \_  ___/  \____/ /_/    \_\ |_|
  \__/ \__/      \ \__
                  \___\ by Red Hat
 Build, Store, and Distribute your Containers
	`)
	return rootCmd.Execute()
}
