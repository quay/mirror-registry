package main

import (
	"context"
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

	// Set up postgres data folder
	log.Printf("Setting up Postgres service\n")
	postgresDataPath := path.Join(installPath, "pg-data")
	err = os.Mkdir(postgresDataPath, 0755)
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
		"POSTGRESQL_USER":            "quay-database",
		"POSTGRESQL_PASSWORD":        "secret-password",
		"POSTGRESQL_ADMIN_PASSWORD":  "secret-password",
		"POSTGRESQL_DATABASE":        "quay-database",
		"POSTGRESQL_SHARED_BUFFERS":  "256MB",
		"POSTGRESQL_MAX_CONNECTIONS": "2000",
	}

	genMounts, _, _, err := specgen.GenVolumeMounts([]string{"/root/quay-install/pg-data:/var/lib/pgsql/data"})
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
