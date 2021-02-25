package cmd

import (
	_ "embed" // embed package is used to embed service files
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

//go:embed "assets/quay.service"
var quayServiceBytes []byte

//go:embed "assets/postgres.service"
var postgresServiceBytes []byte

//go:embed "assets/redis.service"
var redisServiceBytes []byte

type service struct {
	name     string
	image    string
	location string
	bytes    []byte
}

var services = []service{
	{
		"quay-app", "quay.io/projectquay/quay:latest", "/etc/systemd/system/quay-app.service", quayServiceBytes,
	},
	{
		"quay-postgres", "docker.io/centos/postgresql-10-centos8", "/etc/systemd/system/quay-postgres.service", postgresServiceBytes,
	},
	{
		"quay-redis", "docker.io/centos/redis-5-centos8", "/etc/systemd/system/quay-redis.service", redisServiceBytes,
	},
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
		log.Fatalf("An error occurred: %s", err.Error())
	}
}

// verbose is the optional command that will display INFO logs
var verbose bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose logs")
}

var (
	rootCmd = &cobra.Command{
		Use:   "quay-installer",
		Short: "A generator for Cobra based Applications",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.InfoLevel)
			} else {
				log.SetLevel(log.WarnLevel)
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	return rootCmd.Execute()
}
