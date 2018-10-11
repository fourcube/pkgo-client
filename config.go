package main

import (
	"fmt"
	"os"
)

// SystemdConfig is a part of a PkgoConfig
type ProcessManager struct {
	Type    string
	Service string
}

// PkgoConfig describes a package that is updateable with pkgo
type PkgoConfig struct {
	Name           string
	ExecutablePath string `toml:"executable_path"`
	PidFile        string `toml:"pid_file"`
	LicenseKey     string `toml:"license_key"`
}

func (c *PkgoConfig) Validate() (bool, error) {
	if c.PidFile == "" {
		return false, fmt.Errorf("missing pid_file configuration")
	}

	// Check if executable exists
	if !exists(c.ExecutablePath) {
		return false, fmt.Errorf("executable_path %v does not exist", c.ExecutablePath)
	}
	return true, nil
}

// exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
