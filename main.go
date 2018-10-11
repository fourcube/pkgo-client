package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fourcube/pkgo-client/client"
	"github.com/fourcube/pkgo-client/xlog"
)

const SigningKey = `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEZGBXuKuau9Q+cnDCHsN48ovzopce+QcU
qab1BAkJZXNdDHxEoQFnf72TYuzl3LjTsLuIA2tpx55sG79zgJHG6hyso7aUuQ+c
vQrNHMoC/IHD9FkIqWrBH1xZe8LE9X9t
-----END PUBLIC KEY-----`

var (
	Version         = "dev"
	configPathFlag  string
	printVersion    bool
	printSigningKey bool
)

func init() {
	flag.StringVar(&configPathFlag, "c", "pkgo.toml", "path to pkgo.toml")
	flag.BoolVar(&printVersion, "v", false, "print version")
	flag.BoolVar(&printSigningKey, "k", false, "print signature key")
}

func main() {
	flag.Parse()

	if printVersion {
		fmt.Println(Version)
		return
	}

	if printSigningKey {
		fmt.Println(SigningKey)
		return
	}

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
		RepositoryUrl:  config.Repository,
		WorkingDir:     config.WorkingDir,
		ExecutablePath: config.ExecutablePath,
		LicenseKey:     config.LicenseKey,
	})

	if err != nil && err != client.ErrNoUpdates {
		xlog.Fatal("update error: %v", err)
	}

	if err == client.ErrNoUpdates {
		xlog.Print("no updates available")
		return
	}

	if len(config.AfterUpdate) > 0 {
		err = runAfterUpdateCommands(config.AfterUpdate)
	}
	if err != nil {
		xlog.Fatal("after update error: %v", err)
	}

}

func runAfterUpdateCommands(commands []string) error {
	xlog.Print("after_update hooks")
	for _, command := range commands {
		xlog.Print("executing 'command'")
		tokens := strings.Split(command, " ")
		cmd := exec.Command(tokens[0], tokens[1:]...)
		done := make(chan error)

		go func() {
			err := cmd.Run()
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				return err
			}
		case <-time.After(30 * time.Second):
			cmd.Process.Kill()
			xlog.Fatal("timeout during after_update command '%s'", command)
		}

	}
	return nil
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
