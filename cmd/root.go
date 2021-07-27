package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Create logger
var log = &logrus.Logger{
	Out:   os.Stdout,
	Level: logrus.InfoLevel,
}

func watchFileAndRun(filePath string) error {
	t, err := tail.TailFile(filePath, tail.Config{Follow: true})
	check(err)
	for line := range t.Lines {
		if strings.TrimSpace(line.Text) != "" {
			msg := strings.TrimSpace(strings.Split(line.Text, " - ")[2])
			status := strings.TrimSpace(strings.Split(line.Text, " - ")[4])
			if status == "OK" || status == "SKIPPED" {
				log.Info(status + ": " + msg)
			} else {
				log.Error(msg)
			}

		}

	}
	return nil
}

func setupLocalSSH(hostname, username string) error {

	log.Infof("Generating SSH Key")
	cmd := exec.Command("bash", "-c", "ssh-keygen -b 2048 -t rsa -N '' -f ~/.ssh/quay_installer")
	if err := cmd.Run(); err != nil {
		return err
	}
	log.Infof("Generated SSH Key at " + os.Getenv("HOME") + "/.ssh/quay_installer")

	keyFile, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/quay_installer.pub")
	if err != nil {
		return err
	}

	cmd = exec.Command("bash", "-c", "/bin/echo \""+string(keyFile)+"\" >> ~/.ssh/authorized_keys")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
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
		cleanup()
		os.Exit(1)
	}
}

func cleanup() {
	os.RemoveAll("/tmp/app")
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
