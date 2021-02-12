package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"log"
)

func check(err error) {
	if err != nil {
		log.Fatalf("An error occurred: %s", err.Error())
	}
}

func main() {

	log.Printf("Quay All in One Installer\n")

	// Build install path and create directory
	log.Printf("Creating quay-install directory in $HOME\n")
	installPath := path.Join(os.Getenv("HOME"), "quay-install")
	err := os.Mkdir(installPath, 0755)
	check(err)

	log.Printf("Setting up Postgres service\n")
	postgresPath := path.Join(installPath, "pgqsl")
	err = os.Mkdir(postgresPath, 0755)
	check(err)

	output, err := exec.Command("setfacl", "-m", "u:26:-wx", postgresPath).Output()
	check(err)
	fmt.Println(output)

}
