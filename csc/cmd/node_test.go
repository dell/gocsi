package cmd

import (
	"testing"

	"github.com/dell/gocsi/mock/service"
	"github.com/stretchr/testify/assert"
)

func TestNodeCmd(t *testing.T) {
	setupRoot(t)
	err := nodeCmd.PersistentPreRunE(nodeCmd, []string{})
	assert.NoError(t, err)
}

func TestNodeExpandVolumeCmd(t *testing.T) {
	node.client = service.NewClient()
	setupRoot(t)
	child := nodeExpandVolumeCmd
	err := child.RunE(RootCmd, []string{"volID", "/test/volume"})
	assert.NoError(t, err)
}

func TestNodeGetCapabilitiesCmd(t *testing.T) {
	node.client = service.NewClient()
	setupRoot(t)
	child := nodeGetCapabilitiesCmd
	err := child.RunE(RootCmd, []string{})
	assert.NoError(t, err)
}

func TestNodeGetVolumeStatsCmd(t *testing.T) {
	// Set format for NodeGetVolumeStats cmd
	root.format = statsFormat
	setupRoot(t)

	node.client = service.NewClient()
	child := nodeGetVolumeStatsCmd
	err := child.RunE(RootCmd, []string{"Mock Volume 2:/root/mock-vol:/root/mock/patch"})
	assert.NoError(t, err)
}

func TestNodeGetInfo(t *testing.T) {
	// Set format for NodeGetInfo cmd
	root.format = nodeInfoFormat
	setupRoot(t)
	node.client = service.NewClient()
	child := nodeGetInfoCmd
	err := child.RunE(RootCmd, []string{"mock-node-id"})
	assert.NoError(t, err)
}

func TestNodePublishVolume(t *testing.T) {
	setupRoot(t)
	node.client = service.NewClient()
	child := nodePublishVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.NoError(t, err)
}

func TestNodeStageVolume(t *testing.T) {
	setupRoot(t)
	node.client = service.NewClient()
	child := nodeStageVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.NoError(t, err)
}

func TestNodeUnpublishVolume(t *testing.T) {
	setupRoot(t)
	node.client = service.NewClient()
	child := nodeUnpublishVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id", "mock/target/path"})
	assert.NoError(t, err)
}

func TestNodeUnstageVolume(t *testing.T) {
	setupRoot(t)
	node.client = service.NewClient()
	child := nodeUnstageVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id", "mock/target/path"})
	assert.NoError(t, err)
}
