package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestPathExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		if !pathExists(f.Name()) {
			t.Errorf("pathExists(%q) = false, want true", f.Name())
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		if pathExists("/nonexistent/path/file.txt") {
			t.Error("pathExists for nonexistent file = true, want false")
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := t.TempDir()
		if !pathExists(dir) {
			t.Errorf("pathExists(%q) = false, want true for directory", dir)
		}
	})
}

func TestGetImageMetadata(t *testing.T) {
	tests := []struct {
		app         string
		imageName   string
		archivePath string
		wantContain string
	}{
		{
			app:         "pause",
			imageName:   "registry.access.redhat.com/ubi8/pause:8.10-5",
			archivePath: "/tmp/pause.tar",
			wantContain: "registry.access.redhat.com/ubi8/pause:8.10-5",
		},
		{
			app:         "sqlite",
			imageName:   "quay.io/projectquay/sqlite-cli:latest",
			archivePath: "/tmp/sqlite3.tar",
			wantContain: "sqlite3",
		},
		{
			app:         "ansible",
			imageName:   "quay.io/quay/mirror-registry-ee:latest",
			archivePath: "/tmp/ee.tar",
			wantContain: "ansible-runner",
		},
		{
			app:         "redis",
			imageName:   "registry.redhat.io/rhel8/redis-6:1",
			archivePath: "/tmp/redis.tar",
			wantContain: "REDIS_VERSION=6",
		},
		{
			app:         "quay",
			imageName:   "registry.redhat.io/quay/quay-rhel8:v3.12.14",
			archivePath: "/tmp/quay.tar",
			wantContain: "RED_HAT_QUAY=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.app, func(t *testing.T) {
			result := getImageMetadata(tt.app, tt.imageName, tt.archivePath)
			if result == "" {
				t.Fatal("getImageMetadata returned empty string")
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("getImageMetadata(%q, ...) missing %q", tt.app, tt.wantContain)
			}
			if !strings.Contains(result, "/usr/bin/podman image import") {
				t.Errorf("getImageMetadata(%q, ...) missing podman import command", tt.app)
			}
			if !strings.Contains(result, tt.imageName) {
				t.Errorf("getImageMetadata(%q, ...) missing image name %q", tt.app, tt.imageName)
			}
			if !strings.Contains(result, tt.archivePath) {
				t.Errorf("getImageMetadata(%q, ...) missing archive path %q", tt.app, tt.archivePath)
			}
		})
	}

	t.Run("unknown app returns empty", func(t *testing.T) {
		result := getImageMetadata("unknown", "img:latest", "/tmp/archive.tar")
		if result != "" {
			t.Errorf("getImageMetadata for unknown app = %q, want empty", result)
		}
	})
}

// generateTestCertificate creates a self-signed cert/key pair for testing.
// Returns paths to the cert and key files in the given directory.
func generateTestCertificate(t *testing.T, dir, hostname string) (certPath, keyPath string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: hostname},
		DNSNames:     []string{hostname},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	certPath = filepath.Join(dir, "test.cert")
	keyPath = filepath.Join(dir, "test.key")

	certFile, err := os.Create(certPath)
	if err != nil {
		t.Fatal(err)
	}
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certFile.Close()

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyFile, err := os.Create(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyFile.Close()

	return certPath, keyPath
}

func init() {
	log.SetFormatter(&logrus.TextFormatter{})
}

func TestLoadCerts(t *testing.T) {
	t.Run("empty cert and key paths returns nil", func(t *testing.T) {
		err := loadCerts("", "", "example.com", false)
		if err != nil {
			t.Errorf("loadCerts with empty paths returned error: %v", err)
		}
	})

	t.Run("valid cert matching hostname", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := generateTestCertificate(t, dir, "myhost.example.com")

		err := loadCerts(certPath, keyPath, "myhost.example.com", false)
		if err != nil {
			t.Errorf("loadCerts with valid cert returned error: %v", err)
		}
	})

	t.Run("cert hostname mismatch fails", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := generateTestCertificate(t, dir, "correct.example.com")

		err := loadCerts(certPath, keyPath, "wrong.example.com", false)
		if err == nil {
			t.Error("loadCerts with hostname mismatch should return error")
		}
	})

	t.Run("cert hostname mismatch with skipCheck succeeds", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := generateTestCertificate(t, dir, "correct.example.com")

		err := loadCerts(certPath, keyPath, "wrong.example.com", true)
		if err != nil {
			t.Errorf("loadCerts with skipCheck=true returned error: %v", err)
		}
	})

	t.Run("nonexistent cert file fails", func(t *testing.T) {
		dir := t.TempDir()
		_, keyPath := generateTestCertificate(t, dir, "test.example.com")

		err := loadCerts("/nonexistent/cert.pem", keyPath, "test.example.com", true)
		if err == nil {
			t.Error("loadCerts with nonexistent cert should return error")
		}
	})

	t.Run("nonexistent key file fails", func(t *testing.T) {
		dir := t.TempDir()
		certPath, _ := generateTestCertificate(t, dir, "test.example.com")

		err := loadCerts(certPath, "/nonexistent/key.pem", "test.example.com", true)
		if err == nil {
			t.Error("loadCerts with nonexistent key should return error")
		}
	})

	t.Run("malformed cert file fails", func(t *testing.T) {
		dir := t.TempDir()
		certPath := filepath.Join(dir, "bad.cert")
		keyPath := filepath.Join(dir, "bad.key")
		os.WriteFile(certPath, []byte("not a certificate"), 0644)
		os.WriteFile(keyPath, []byte("not a key"), 0644)

		err := loadCerts(certPath, keyPath, "test.example.com", false)
		if err == nil {
			t.Error("loadCerts with malformed cert should return error")
		}
	})
}

func TestIsLocalInstall(t *testing.T) {
	// Save and restore package-level vars
	origHostname := targetHostname
	origUsername := targetUsername
	defer func() {
		targetHostname = origHostname
		targetUsername = origUsername
	}()

	t.Run("localhost is always local", func(t *testing.T) {
		targetHostname = "localhost"
		targetUsername = "someuser"
		if !isLocalInstall() {
			t.Error("isLocalInstall() = false for localhost, want true")
		}
	})

	t.Run("remote hostname is not local", func(t *testing.T) {
		targetHostname = "remote.example.com"
		targetUsername = os.Getenv("USER")
		if isLocalInstall() {
			t.Error("isLocalInstall() = true for remote host, want false")
		}
	})
}
