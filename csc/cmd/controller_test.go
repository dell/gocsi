package cmd

import (
	"testing"
)

func TestControllerCmd(t *testing.T) {
	child, _, err := RootCmd.Find([]string{"controller"})
	if err != nil {
		t.Errorf("Unable to find cmd")
	}
	child.PersistentPreRunE(RootCmd, []string{})

	// TODO: Add test case for error condition in PreRun
}
