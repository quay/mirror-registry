package cmd

import (
	"crypto/tls"
	"crypto/x509"
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
	statement := getImageMetadata("ansible", eeImage, executionEnvironmentPath)
	cmd := exec.Command("/bin/bash", "-c", statement)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	log.Debug("Importing execution enviornment with command: ", cmd)

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

func loadCerts(certFile, keyFile, hostname string, skipCheck bool) error {
	if certFile != "" && keyFile != "" {
		log.Info("Loading SSL certificate file " + certFile)
		log.Info("Loading SSL key file " + keyFile)
		if !skipCheck {
			certKey, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				log.Errorf("Failed loading certificate and key file: %s", err.Error())
				return err
			}

			cert, err := x509.ParseCertificate(certKey.Certificate[0])
			if err != nil {
				log.Errorf("Failed parsing certificate file: %s", err.Error())
				return err
			}

			roots := x509.NewCertPool()
			// Allow self-signed certificate and do not check the issuer
			roots.AddCert(cert)

			opts := x509.VerifyOptions{
				DNSName:   hostname,
				Roots:     roots,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			}

			_, err = cert.Verify(opts)
			if err != nil {
				log.Errorf("Failed verifying certificate: %s", err.Error())
				return err
			}
			log.Info("SSL certificate check succeeded")
		}

		if pathExists(certFile) {
			setSELinux(certFile)
		} else {
			return errors.New("Certificate file not found: " + certFile)
		}

		if pathExists(keyFile) {
			setSELinux(keyFile)
		} else {
			return errors.New("Certificate key file not found: " + keyFile)
		}
	}

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
	return !os.IsNotExist(err)
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
		Use: "mirror-registry",
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
