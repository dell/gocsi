package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControllerCmd(t *testing.T) {
	child := controllerCmd

	// test case: no error
	err := child.PersistentPreRunE(child, []string{})
	assert.NoError(t, err)

	// TODO: Add test case for error condition in PreRun
	// The only way to do that is to trigger an error in root command's
	// prerun, and I don't see any args that can be passed to it that would cause such a thing
	// Very high effort for one line of coverage.

}

func TestCreateSnapshotCmd(t *testing.T) {
	child := createSnapshotCmd

	// TODO: Valid test cases
	child.RunE(RootCmd, []string{"name"})

	// TODO: Add test case for error condition in RunE
}
