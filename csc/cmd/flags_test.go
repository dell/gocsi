package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_flagWithSuccessAlreadyExists(t *testing.T) {
	child := createVolumeCmd
	var withSuccessAlreadyExists bool

	flagWithSuccessAlreadyExists(child.Flags(), &withSuccessAlreadyExists, "false")

	// ensure the flag was added
	assert.NotEqual(t, child.Flags().Lookup("with-success-create-already-exists"), nil)
}

func Test_flagWithSuccessNotFound(t *testing.T) {
	child := deleteVolumeCmd
	var withSuccessNotFound bool

	flagWithSuccessNotFound(child.Flags(), &withSuccessNotFound, "false")

	// ensure the flag was added
	assert.NotEqual(t, child.Flags().Lookup("with-success-not-found"), nil)
}
