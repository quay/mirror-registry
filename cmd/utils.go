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
	"strings"
)

// This variable is set at build time via ldflags
var sqliteImage string

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
	log.Debug("Importing execution environment with command: ", cmd)

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func isLocalInstall() bool {
	if targetHostname == "localhost" || targetHostname == getFQDN() && targetUsername == os.Getenv("USER") {
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
	cmd = exec.Command("bash", "-c", "umask 066 && /bin/echo \""+string(keyFile)+"\" >> ~/.ssh/authorized_keys")
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

func loadSqliteCli() (string, error) {
	// Ensure execution environment is present
	executableDir, err := os.Executable()
	if err != nil {
		return "", err
	}
	sqliteArchivePath := path.Join(path.Dir(executableDir), "sqlite3.tar")
	if !pathExists(sqliteArchivePath) {
		return "", errors.New("Could not find sqlite3.tar at " + sqliteArchivePath)
	}
	log.Info("Found sqlite3 cli binary at " + sqliteArchivePath)

	sqliteArchiveMountFlag := fmt.Sprintf(" -v %s:/runner/sqlite3.tar", sqliteArchivePath)

	if isLocalInstall() {
		// Load sqlite3 as a podman image
		log.Printf("Loading sqlite3 cli binary from sqlite3.tar")
		statement := getImageMetadata("sqlite", sqliteImage, sqliteArchivePath)
		sqliteImportCmd := exec.Command("/bin/bash", "-c", statement)
		if verbose {
			sqliteImportCmd.Stderr = os.Stderr
			sqliteImportCmd.Stdout = os.Stdout
		}
		log.Debug("Importing sqlite3 cli binary with command: ", sqliteImportCmd)
		err = sqliteImportCmd.Run()
		if err != nil {
			return "", err
		}
	}
	log.Infof("Attempting to set SELinux rules on sqlite archive")
	cmd := exec.Command("chcon", "-Rt", "svirt_sandbox_file_t", sqliteArchivePath)
	if verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		log.Warn("Could not set SELinux rule. If your system does not have SELinux enabled, you may ignore this.")
	}
	return sqliteArchiveMountFlag, nil
}

// getImageMetadata provides the metadata needed for a corresponding image
func getImageMetadata(app, imageName, archivePath string) string {
	var statement string

	switch app {
	case "pause":
		statement = `/usr/bin/podman image import \
					--change 'ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV container=oci' \
					--change 'ENTRYPOINT=["sleep"]' \
					--change 'CMD=["infinity"]' \
					- ` + imageName + ` < ` + archivePath
	case "sqlite":
		statement = `/usr/bin/podman image import \
					--change 'ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV container=oci' \
					--change 'ENTRYPOINT=["/usr/bin/sqlite3"]' \
					- ` + imageName + ` < ` + archivePath
	case "ansible":
		statement = `/usr/bin/podman image import \
					--change 'ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV HOME=/home/runner' \
					--change 'ENV container=oci' \
					--change 'ENTRYPOINT=["entrypoint"]' \
					--change 'WORKDIR=/runner' \
					--change 'EXPOSE=6379' \
					--change 'VOLUME=/runner' \
					--change 'CMD ["ansible-runner", "run", "/runner"]' \
					- ` + imageName + ` < ` + archivePath
	case "redis":
		statement = `/usr/bin/podman image import \
					--change 'ENV PATH=/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV container=oci' \
					--change 'ENV STI_SCRIPTS_URL=image:///usr/libexec/s2i' \
					--change 'ENV STI_SCRIPTS_PATH=/usr/libexec/s2i' \
					--change 'ENV APP_ROOT=/opt/app-root' \
					--change 'ENV HOME=/var/lib/redis' \
					--change 'ENV PLATFORM=el8' \
					--change 'ENV REDIS_VERSION=6' \
					--change 'ENV CONTAINER_SCRIPTS_PATH=/usr/share/container-scripts/redis' \
					--change 'ENV REDIS_PREFIX=/usr' \
					--change 'ENV REDIS_CONF=/etc/redis.conf' \
					--change 'ENTRYPOINT=["container-entrypoint"]' \
					--change 'USER=1001' \
					--change 'WORKDIR=/opt/app-root/src' \
					--change 'EXPOSE=6379' \
					--change 'VOLUME=/var/lib/redis/data' \
					--change 'CMD ["run-redis"]' \
					- ` + imageName + ` < ` + archivePath
	case "quay":
		// quay.io
		statement = `/usr/bin/podman image import \
					--change 'ENV container=oci' \
					--change 'ENV PATH=/opt/app-root/bin:/opt/app-root/src/.local/bin:/opt/app-root/src/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV PYTHONUNBUFFERED=1' \
					--change 'ENV PYTHONIOENCODING=UTF-8' \
					--change 'ENV LC_ALL=C.UTF-8' \
					--change 'ENV LANG=C.UTF-8' \
					--change 'ENV QUAYDIR=/quay-registry' \
					--change 'ENV QUAYCONF=/quay-registry/conf' \
					--change 'ENV QUAYRUN=/quay-registry/conf' \
					--change 'ENV QUAYPATH=/quay-registry' \
					--change 'ENV PYTHONUSERBASE=/app' \
					--change 'ENV PYTHONPATH=/quay-registry' \
					--change 'ENV TZ=UTC' \
					--change 'ENV RED_HAT_QUAY=true' \
					--change 'ENTRYPOINT=["dumb-init","--","/quay-registry/quay-entrypoint.sh"]' \
					--change 'WORKDIR=/quay-registry' \
					--change 'EXPOSE=7443' \
					--change 'EXPOSE=8080' \
					--change 'EXPOSE=8443' \
					--change 'VOLUME=/conf/stack' \
					--change 'VOLUME=/datastorage' \
					--change 'VOLUME=/sqlite' \
					--change 'VOLUME=/tmp' \
					--change 'VOLUME=/var/log' \
					--change 'USER=1001' \
					--change 'CMD ["registry"]' \
					- ` + imageName + ` < ` + archivePath
	}

	return statement
}

// checkInput validates user input against available options
func getApproval(question string) bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y":
		return true
	case "n":
		return false
	default:
		fmt.Println("Invalid input.", question)
		return getApproval(question)
	}
}

func getFQDN() string {
	fqdn, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		errorMessage := "Failed to automatically acquire host FQDN, please set manually with --targetHostname. "
		log.Fatal(errorMessage, err)
	}

	return strings.TrimSuffix(string(fqdn), "\n")
}
