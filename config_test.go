package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidateMissingPidFile(t *testing.T) {
	c := PkgoConfig{}

	valid, err := c.Validate()
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestConfigValidateMissingExecutable(t *testing.T) {
	c := PkgoConfig{
		ExecutablePath: "/tmp/foo",
	}

	valid, err := c.Validate()
	assert.NotNil(t, err)
	assert.False(t, valid)
}
