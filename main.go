package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"log"

	"github.com/containers/podman/v2/libpod/define"
	"github.com/containers/podman/v2/pkg/bindings"
	"github.com/containers/podman/v2/pkg/bindings/containers"
	"github.com/containers/podman/v2/pkg/bindings/images"
	"github.com/containers/podman/v2/pkg/domain/entities"
	"github.com/containers/podman/v2/pkg/specgen"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	quayImage     = "quay.io/projectquay/quay"
	postgresImage = "docker.io/centos/postgresql-10-centos7:latest"
)

func check(err error) {
	if err != nil {
		log.Fatalf("An error occurred: %s", err.Error())
	}
}

func main() {

	log.Printf("Quay All in One Installer\n")

	// Start service
	log.Printf("Checking for Podman socket")
	_, err := exec.Command("systemctl", "start", "podman.socket").Output()
	check(err)

	// Connect to socket
	connText, err := bindings.NewConnection(context.Background(), "unix://run/podman/podman.sock")
	check(err)

	// Build install path and create directory
	log.Printf("Creating quay-install directory in $HOME\n")
	installPath := path.Join(os.Getenv("HOME"), "quay-install")
	err = os.Mkdir(installPath, 0755)
	check(err)

	// Setup Postgres Container
	setupPostgres(connText, installPath)

	// Setup Quay Container
	setupQuay(connText, installPath)

}

func setupPostgres(connText context.Context, installPath string) {

	log.Printf("Setting up Postgres service\n")

	// Set up postgres data folder
	postgresDataPath := path.Join(installPath, "pg-data")
	err := os.Mkdir(postgresDataPath, 0755)
	check(err)

	_, err = exec.Command("setfacl", "-m", "u:26:-wx", postgresDataPath).Output()
	check(err)

	// Pull postgres image
	log.Printf("Pulling Postgres image")
	_, err = images.Pull(connText, postgresImage, entities.ImagePullOptions{})
	check(err)

	// Create postgres container spec
	s := specgen.NewSpecGenerator(postgresImage, false)
	s.Terminal = true
	s.PortMappings = []specgen.PortMapping{
		{
			HostPort:      5432,
			ContainerPort: 5432,
		},
	}
	s.Env = map[string]string{
		"POSTGRESQL_USER":            "user",
		"POSTGRESQL_PASSWORD":        "password",
		"POSTGRESQL_ADMIN_PASSWORD":  "password",
		"POSTGRESQL_DATABASE":        "quay-database",
		"POSTGRESQL_SHARED_BUFFERS":  "256MB",
		"POSTGRESQL_MAX_CONNECTIONS": "2000",
	}

	genMounts, _, _, err := specgen.GenVolumeMounts([]string{fmt.Sprintf("%s:/var/lib/pgsql/data", postgresDataPath)})
	s.Mounts = []specs.Mount{
		genMounts["/var/lib/pgsql/data"],
	}

	err = s.Validate()
	check(err)

	// Create container
	log.Printf("Creating Postgres container")
	r, err := containers.CreateWithSpec(connText, s)
	check(err)

	// Start container and wait for start
	log.Printf("Starting Postgres container")
	err = containers.Start(connText, r.ID, nil)
	check(err)
	running := define.ContainerStateRunning
	_, err = containers.Wait(connText, r.ID, &running)
	check(err)
}

func setupQuay(connText context.Context, installPath string) {

	log.Printf("Setting up Quay service\n")

	// Build Quay Config
	quayConfigPath := path.Join(installPath, "quay-config")
	err := os.Mkdir(quayConfigPath, 0755)
	check(err)
	configBytes, err := createConfigBytes()
	check(err)
	err = ioutil.WriteFile(path.Join(quayConfigPath, "config.yaml"), configBytes, 0644)
	check(err)

	// Set up Quay local storage
	quayStoragePath := path.Join(installPath, "quay-storage")
	err = os.Mkdir(quayStoragePath, 0755)
	check(err)
	_, err = exec.Command("setfacl", "-m", "u:1001:-wx", quayStoragePath).Output()
	check(err)

	// Pull postgres image
	log.Printf("Pulling Quay image")
	_, err = images.Pull(connText, quayImage, entities.ImagePullOptions{})
	check(err)

	// Create postgres container spec
	s := specgen.NewSpecGenerator(quayImage, false)
	s.Terminal = true
	s.PortMappings = []specgen.PortMapping{
		{
			HostPort:      8080,
			ContainerPort: 8080,
		},
	}

	genMounts, _, _, err := specgen.GenVolumeMounts([]string{fmt.Sprintf("%s:/conf/stack", quayConfigPath), fmt.Sprintf("%s:/datastorage", quayStoragePath)})
	s.Mounts = []specs.Mount{
		genMounts["/conf/stack"],
		genMounts["/datastorage"],
	}

	err = s.Validate()
	check(err)

	// Create container
	log.Printf("Creating Quay container")
	r, err := containers.CreateWithSpec(connText, s)
	check(err)

	// Start container and wait for start
	log.Printf("Starting Quay container")
	err = containers.Start(connText, r.ID, nil)
	check(err)
	running := define.ContainerStateRunning
	_, err = containers.Wait(connText, r.ID, &running)
	check(err)

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
  password: strongpassword
  port: 6379
DATABASE_SECRET_KEY: "81541057085600720484162638317561463611194901378275494293746615390984668417511"
DB_URI: postgresql://user:password@localhost/quay-database
DEFAULT_TAG_EXPIRATION: 2w
DISTRIBUTED_STORAGE_DEFAULT_LOCATIONS: []
DISTRIBUTED_STORAGE_PREFERENCE:
  - localstorage
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
MAIL_USERNAME: jonathan
MAIL_PASSWORD: king
MAIL_USE_AUTH: true
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
MAIL_DEFAULT_SENDER: support@quay.io
MAIL_PORT: 587
MAIL_USE_TLS: true
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
  - user
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
  host: 192.168.250.159
  password: strongpassword
  port: 6379
USE_CDN: false`)

	return configBytes, nil

}
