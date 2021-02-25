package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// imageArchiveDir is the optional location of the OCI image archive containing required install images
var imageArchiveDir string

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

	// Add --config-dir flag
	installCmd.Flags().StringVarP(&imageArchiveDir, "image-archive", "i", "", "An archive containing images")

	// // Add --password flag
	// editorCmd.Flags().StringVarP(&editorPassword, "password", "p", "", "The password to enter the editor")
	// editorCmd.MarkFlagRequired("password")

}

func install() {
	log.Printf("Installing Quay")

	var err error
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	// Build install path and create directory
	installPath := path.Join(os.Getenv("HOME"), "quay-install")
	log.Printf("Creating quay-install directory at %s\n", installPath)
	err = os.Mkdir(installPath, 0755)
	check(err)

	// Build pg-data directory for postgres and set permissions
	postgresDataPath := path.Join(installPath, "pg-data")
	log.Printf("Creating pg-data in %s", postgresDataPath)
	err = os.Mkdir(postgresDataPath, 0755)
	check(err)
	_, err = exec.Command("setfacl", "-m", "u:26:-wx", postgresDataPath).Output()
	check(err)

	// Build quay-storage directory for Quay local storage and set permissions
	quayStoragePath := path.Join(installPath, "quay-storage")
	log.Printf("Creating quay-storage directory at %s\n", quayStoragePath)
	err = os.Mkdir(quayStoragePath, 0755)
	check(err)
	_, err = exec.Command("setfacl", "-m", "u:1001:-wx", quayStoragePath).Output()
	check(err)

	// Build quay config path and write out
	quayConfigPath := path.Join(installPath, "quay-config")
	log.Printf("Creating quay-config directory at %s\n", quayConfigPath)
	err = os.Mkdir(quayConfigPath, 0755)
	check(err)
	configBytes, err := createConfigBytes()
	check(err)
	err = ioutil.WriteFile(path.Join(quayConfigPath, "config.yaml"), configBytes, 0644)
	check(err)

	// If SELinux is enabled, set rule
	log.Infof("Setting SELinux rules.")
	cmd := exec.Command("sudo", "setsebool", "-P", "container_manage_cgroup", "1")
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	err = cmd.Run()
	if err != nil {
		if !strings.Contains(stdErr.String(), "command not found") {
			check(errors.New(stdErr.String()))
		}
		log.Warnf("SELinux not found. Skipping.")
	}

	// If image archive is set, load images. Otherwise, pull from dockerhub.
	if imageArchiveDir == "" { // Flag not set
		// Attempt to autodetect
		executableDir, err := os.Executable()
		if err != nil {
			check(err)
		}
		defaultArchive := path.Join(path.Dir(executableDir), "image-archive.tar")
		if pathExists(defaultArchive) { // Autodetect found archive in same dir as executable
			log.Printf("Autodetected image archive at %s", defaultArchive)
			cmd := exec.Command("podman", "load", "-i", defaultArchive)
			fmt.Print("\033[34m")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				check(errors.New(stdErr.String()))
			}
			fmt.Print("\033[0m")
		} else { // No archive provided, pulling images automatically
			log.Printf("Pulling required images")
			for _, s := range services {
				cmd := exec.Command("podman", "pull", s.image)
				fmt.Print("\033[34m")
				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
				err = cmd.Run()
				if err != nil {
					check(errors.New(stdErr.String()))
				}
				fmt.Print("\033[0m")
			}
		}
	} else { // Flag was set
		if pathExists(imageArchiveDir) { // Autodetect found archive in same dir as executable
			log.Printf("Using specified image archive at %s", imageArchiveDir)
			cmd := exec.Command("podman", "load", "-i", imageArchiveDir)
			fmt.Print("\033[34m")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				check(errors.New(stdErr.String()))
			}
			fmt.Print("\033[0m")
		}
	}

	// Write systemd files and enable service
	for _, s := range services {
		log.Printf("Writing unit file for %s in %s", s.name, s.location)
		err = ioutil.WriteFile(s.location, s.bytes, 0644)
		check(err)
		cmd := exec.Command("systemctl", "enable", s.name)
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err = cmd.Run()
		if err != nil {
			check(errors.New(stdErr.String()))
		}
		cmd = exec.Command("systemctl", "start", s.name)
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err = cmd.Run()
		if err != nil {
			check(errors.New(stdErr.String()))
		}
		log.Printf("Successfully set up %s", s.name)
	}

	// Install trgm and create user in database
	time.Sleep(10 * time.Second)
	log.Printf("Installing pg_trgm in Postgres")
	cmd = exec.Command("podman", "exec", "-it", "quay-postgresql-service", "/bin/bash", "-c", "echo \"CREATE EXTENSION IF NOT EXISTS pg_trgm\" | psql -d quay -U postgres")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		check(err)
	}

	// Reload
	_, err = exec.Command("systemctl", "daemon-reexec").Output()
	check(err)
	log.Printf("Quay installed successfully")
}

func createConfigBytes() ([]byte, error) {

	// FIX THIS
	// Create base Quay config
	// options := generate.AioiInputOptions{
	// 	DatabaseURI:    "postgresql://user:password@localhost:5432/quay-database",
	// 	ServerHostname: "localhost:8080",
	// 	RedisHostname:  "localhost",
	// 	RedisPassword:  "strong-password",
	// }
	// config, err := generate.GenerateBaseConfig(options)
	// check(err)

	// configBytes, err := yaml.Marshal(config)
	// check(err)
	// fmt.Println(string(configBytes))

	configBytes := []byte(`AUTHENTICATION_TYPE: Database
BUILDLOGS_REDIS:
  host: localhost
  password: password
  port: 6379
DATABASE_SECRET_KEY: "81541057085600720484162638317561463611194901378275494293746615390984668417511"
DB_URI: postgresql://user:password@localhost/quay
DEFAULT_TAG_EXPIRATION: 2w
DISTRIBUTED_STORAGE_DEFAULT_LOCATIONS: []
DISTRIBUTED_STORAGE_PREFERENCE:
  - default
DISTRIBUTED_STORAGE_CONFIG:
  default:
    - LocalStorage
    - storage_path: /datastorage
ENTERPRISE_LOGO_URL: /static/img/quay-horizontal-color.svg
FEATURE_ACI_CONVERSION: false
FEATURE_ANONYMOUS_ACCESS: true
FEATURE_APP_REGISTRY: false
FEATURE_APP_SPECIFIC_TOKENS: true
FEATURE_BUILD_SUPPORT: false
FEATURE_CHANGE_TAG_EXPIRATION: true
FEATURE_DIRECT_LOGIN: true
FEATURE_PARTIAL_USER_AUTOCOMPLETE: true
FEATURE_REPO_MIRROR: false
FEATURE_MAILING: false
FEATURE_REQUIRE_TEAM_INVITE: true
FEATURE_RESTRICTED_V1_PUSH: true
FEATURE_SECURITY_NOTIFICATIONS: true
FEATURE_SECURITY_SCANNER: false
FEATURE_USERNAME_CONFIRMATION: true
FEATURE_USER_CREATION: true
FEATURE_USER_LOG_ACCESS: true
GITHUB_LOGIN_CONFIG: {}
GITHUB_TRIGGER_CONFIG: {}
GITLAB_TRIGGER_KIND: {}
LOGS_MODEL: database
LOGS_MODEL_CONFIG: {}
LOG_ARCHIVE_LOCATION: default
PREFERRED_URL_SCHEME: http
REGISTRY_TITLE: Red Hat Quay
REGISTRY_TITLE_SHORT: Red Hat Quay
REPO_MIRROR_SERVER_HOSTNAME: null
REPO_MIRROR_TLS_VERIFY: true
SECRET_KEY: "30824339799025335633887256663000123118247018465144108496567331049820667127217"
SECURITY_SCANNER_ISSUER_NAME: security_scanner
SERVER_HOSTNAME: quay
SETUP_COMPLETE: true
SUPER_USERS:
  - admin
TAG_EXPIRATION_OPTIONS:
  - 0s
  - 1d
  - 1w
  - 2w
  - 4w
TEAM_RESYNC_STALE_TIME: 60m
TESTING: false
USERFILES_LOCATION: default
USERFILES_PATH: userfiles/
USER_EVENTS_REDIS:
  host: localhost
  password: password
  port: 6379
USE_CDN: false`)

	return configBytes, nil

}
