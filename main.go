package main

import (
	"encoding/json"
	"flag"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fourcube/pkgo-client/client"
	"github.com/fourcube/pkgo-client/xlog"
)

var (
	configPathFlag string
)

func init() {
	flag.StringVar(&configPathFlag, "c", "pkgo.toml", "path to pkgo.toml")
}

func main() {
	flag.Parse()

	configPath, err := filepath.Abs(configPathFlag)
	if err != nil {
		xlog.Fatal("Failed to build absolute path to %s: %v", configPathFlag, err)
	}

	xlog.Print("pkgo config: %s", configPath)
	config := PkgoConfig{}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		xlog.Fatal("failed to read config %v", err)
	}

	if ok, err := config.Validate(); !ok || err != nil {
		xlog.Fatal("Error: %v", err)
	}

	version := findCurrentVersion(config)
	publicKey := findPublicKey(config)

	xlog.Print("updating: %s", config.ExecutablePath)
	xlog.Print("current version: %s", version)

	err = client.Update(client.Package{
		Name:           "openiban",
		PublicKey:      publicKey,
		CurrentVersion: version,
		ExecutablePath: config.ExecutablePath,
		LicenseKey:     config.LicenseKey,
	})

	if err != nil {
		xlog.Print("error: %v", err)
	}
}

func findCurrentVersion(config PkgoConfig) string {
	printVersionCmd := exec.Command(config.ExecutablePath, "-v")

	out, err := printVersionCmd.CombinedOutput()
	if err != nil {
		xlog.Fatal("failed to get output %v", err)
	}

	return strings.TrimSpace(string(out))
}

func findPublicKey(config PkgoConfig) string {
	printVersionCmd := exec.Command(config.ExecutablePath, "-k")

	out, err := printVersionCmd.CombinedOutput()
	if err != nil {
		xlog.Fatal("failed to get output %v", err)
	}

	return strings.TrimSpace(string(out))
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
