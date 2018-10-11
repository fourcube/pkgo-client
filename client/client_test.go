package client_test

import (
	"testing"

	client "github.com/fourcube/pkgo-client/client"
)

func TestUpdate(t *testing.T) {
	client.Update("foo")
}
