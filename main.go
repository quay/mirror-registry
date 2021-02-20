package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"log"
)

const (
	postgresSystemD = `
[Unit]
Description=Podman container-brave_noether.service
Documentation=man:podman-generate-systemd(1)
Wants=network.target
After=network-online.target

[Service]
Environment=PODMAN_SYSTEMD_UNIT=%n
Restart=always
ExecStartPre=/bin/rm -f %t/container-brave_noether.pid %t/container-brave_noether.ctr-id
ExecStart=/usr/bin/podman run --conmon-pidfile %t/container-brave_noether.pid --cidfile %t/container-brave_noether.ctr-id --cgroups=no-conmon -p 5432:5432 -e POSTGRESQL_USER=quay-database -e POSTGRESQL_DATABASE=quay-database -e POSTGRESQL_PASSWORD=quay-database -e POSTGRESQL_ADMIN_PASSWORD=postgres -e POSTGRESQL_SHARED_BUFFERS=256MB -e POSTGRESQL_MAX_CONNECTIONS=2000 -v $HOME/quay-install/pgsql:/var/lib/pgsql/data centos/postgresql-10-centos7@sha256:de1560cb35e5ec643e7b3a772ebaac8e3a7a2a8e8271d9e91ff023539b4dfb33
ExecStop=/usr/bin/podman stop --ignore --cidfile %t/container-brave_noether.ctr-id -t 10
ExecStopPost=/usr/bin/podman rm --ignore -f --cidfile %t/container-brave_noether.ctr-id
PIDFile=%t/container-brave_noether.pid
KillMode=none
Type=forking

[Install]
WantedBy=multi-user.target default.target
`
)

func check(err error) {
	if err != nil {
		log.Fatalf("An error occurred: %s", err.Error())
	}
}

func main() {

	log.Printf("Quay All in One Installer\n")

	// Start service FIXME (just checking for installation)
	log.Printf("Checking for Podman socket")
	_, err := exec.Command("systemctl", "start", "podman.socket").Output()
	check(err)

	// Build install path and create directory
	log.Printf("Creating quay-install directory in $HOME\n")
	installPath := path.Join(os.Getenv("HOME"), "quay-install")
	err = os.Mkdir(installPath, 0755)
	check(err)

	// Build postgres directory
	postgresDataPath := path.Join(installPath, "pg-data")
	err = os.Mkdir(postgresDataPath, 0755)
	check(err)

	// Set permissions
	_, err = exec.Command("setfacl", "-m", "u:26:-wx", postgresDataPath).Output()
	check(err)
	_, err = exec.Command("chcon", "-Rt", "svirt_sandbox_file_t", postgresDataPath).Output()
	check(err)

	log.Printf("Setting up Quay service\n")
	quayConfigPath := path.Join(installPath, "quay-config")
	err = os.Mkdir(quayConfigPath, 0755)
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
	_, err = exec.Command("chcon", "-Rt", "svirt_sandbox_file_t", quayStoragePath).Output()
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
