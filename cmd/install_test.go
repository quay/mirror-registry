package cmd

import (
	"os"
	"testing"
)

func TestInstallFlagDefaults(t *testing.T) {
	flag := installCmd.Flags()

	tests := []struct {
		name     string
		flagName string
		want     string
	}{
		{"initUser default", "initUser", "init"},
		{"initPassword default", "initPassword", ""},
		{"quayHostname default", "quayHostname", ""},
		{"imageArchive default", "image-archive", ""},
		{"sslCert default", "sslCert", ""},
		{"sslKey default", "sslKey", ""},
		{"quayRoot default", "quayRoot", "~/quay-install"},
		{"quayStorage default", "quayStorage", "quay-storage"},
		{"sqliteStorage default", "sqliteStorage", "sqlite-storage"},
		{"additionalArgs default", "additionalArgs", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flag.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("flag %q not registered on install command", tt.flagName)
			}
			if f.DefValue != tt.want {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, f.DefValue, tt.want)
			}
		})
	}
}

func TestInstallBoolFlagDefaults(t *testing.T) {
	flag := installCmd.Flags()

	tests := []struct {
		name     string
		flagName string
		want     string
	}{
		{"sslCheckSkip default", "sslCheckSkip", "false"},
		{"askBecomePass default", "askBecomePass", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flag.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("flag %q not registered on install command", tt.flagName)
			}
			if f.DefValue != tt.want {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, f.DefValue, tt.want)
			}
		})
	}
}

func TestInstallFlagShorthands(t *testing.T) {
	flag := installCmd.Flags()

	tests := []struct {
		flagName  string
		shorthand string
	}{
		{"targetHostname", "H"},
		{"targetUsername", "u"},
		{"ssh-key", "k"},
		{"image-archive", "i"},
		{"quayRoot", "r"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			f := flag.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("flag %q not registered", tt.flagName)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, f.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestUpgradeFlagDefaults(t *testing.T) {
	flag := upgradeCmd.Flags()

	tests := []struct {
		name     string
		flagName string
		want     string
	}{
		{"quayHostname default", "quayHostname", ""},
		{"quayRoot default", "quayRoot", "~/quay-install"},
		{"quayStorage default", "quayStorage", "quay-storage"},
		{"sqliteStorage default", "sqliteStorage", "sqlite-storage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flag.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("flag %q not registered on upgrade command", tt.flagName)
			}
			if f.DefValue != tt.want {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, f.DefValue, tt.want)
			}
		})
	}
}

func TestUninstallFlagDefaults(t *testing.T) {
	flag := uninstallCmd.Flags()

	tests := []struct {
		name     string
		flagName string
		want     string
	}{
		{"targetHostname default", "targetHostname", "localhost"},
		{"quayRoot default", "quayRoot", "~/quay-install"},
		{"quayStorage default", "quayStorage", "quay-storage"},
		{"sqliteStorage default", "sqliteStorage", "sqlite-storage"},
		{"autoApprove default", "autoApprove", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flag.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("flag %q not registered on uninstall command", tt.flagName)
			}
			if f.DefValue != tt.want {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, f.DefValue, tt.want)
			}
		})
	}
}

func TestSshKeyDefault(t *testing.T) {
	homeDir := os.Getenv("HOME")
	expected := homeDir + "/.ssh/quay_installer"

	f := installCmd.Flags().Lookup("ssh-key")
	if f == nil {
		t.Fatal("ssh-key flag not registered on install command")
	}
	if f.DefValue != expected {
		t.Errorf("ssh-key default = %q, want %q", f.DefValue, expected)
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	cmds := rootCmd.Commands()
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name()] = true
	}

	for _, want := range []string{"install", "upgrade", "uninstall"} {
		if !names[want] {
			t.Errorf("root command missing subcommand %q", want)
		}
	}
}

func TestGlobalFlags(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("verbose")
	if f == nil {
		t.Fatal("verbose flag not registered")
	}
	if f.Shorthand != "v" {
		t.Errorf("verbose shorthand = %q, want %q", f.Shorthand, "v")
	}

	f = rootCmd.PersistentFlags().Lookup("no-color")
	if f == nil {
		t.Fatal("no-color flag not registered")
	}
	if f.Shorthand != "c" {
		t.Errorf("no-color shorthand = %q, want %q", f.Shorthand, "c")
	}
}
