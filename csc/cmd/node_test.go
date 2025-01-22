package cmd

import (
	"context"
	"testing"

	"github.com/dell/gocsi/mock/service"
	"github.com/stretchr/testify/assert"
)

func TestNodeCmd(t *testing.T) {
	root.ctx = context.Background()
	err := nodeCmd.PersistentPreRunE(nodeCmd, []string{})
	assert.NoError(t, err)
}

func TestNodeExpandVolumeCmd(t *testing.T) {
	node.client = service.NewClient()
	root.ctx = context.Background()

	child := nodeExpandVolumeCmd
	err := child.RunE(RootCmd, []string{"1", "2"})
	assert.NoError(t, err)
}
