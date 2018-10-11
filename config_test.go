package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidateMissingExecutable(t *testing.T) {
	c := PkgoConfig{
		PidFile:        "/tmp/pid",
		WorkingDir:     "/tmp",
		ExecutablePath: "/tmp/foo",
	}

	valid, err := c.Validate()
	assert.Equal(t, ErrExecutableDoesNotExist, err)
	assert.False(t, valid)
}

func TestConfigValidateMissingWorkingDir(t *testing.T) {
	c := PkgoConfig{
		PidFile:        "/tmp/pid",
		ExecutablePath: "/tmp/foo",
	}

	valid, err := c.Validate()
	assert.Equal(t, ErrMissingWorkingDir, err)
	assert.False(t, valid)
}
