package main

import (
	"errors"
	"os"
)

var (
	ErrMissingWorkingDir      = errors.New("work_dir missing from config")
	ErrExecutableDoesNotExist = errors.New("executable_path does not exist")
)

// SystemdConfig is a part of a PkgoConfig
type ProcessManager struct {
	Type    string
	Service string
}

type Verify struct {
}

// PkgoConfig describes a package that is updateable with pkgo
type PkgoConfig struct {
	Name           string
	Repository     string   `toml:"repository"`
	WorkingDir     string   `toml:"working_dir"`
	ExecutablePath string   `toml:"executable_path"`
	PidFile        string   `toml:"pid_file"`
	LicenseKey     string   `toml:"license_key"`
	AfterUpdate    []string `toml:"after_update"`
}

func (c *PkgoConfig) Validate() (bool, error) {
	if c.WorkingDir == "" {
		return false, ErrMissingWorkingDir
	}

	// Check if executable exists
	if !exists(c.ExecutablePath) {
		return false, ErrExecutableDoesNotExist
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
