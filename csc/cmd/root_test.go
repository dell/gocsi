package cmd

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSettingFormat(t *testing.T) {
	// set root.format blank, verify function sets it correctly bssed on cmd and paging
	listVolumes.paging = false
	root.format = ""
	RootCmd.PersistentPreRunE(listVolumesCmd, []string{})
	assert.Equal(t, root.format, listVolumesFormat)

	listVolumes.paging = true
	root.format = ""
	RootCmd.PersistentPreRunE(listVolumesCmd, []string{})
	assert.Equal(t, root.format, volumeInfoFormat)

	listSnapshots.paging = true
	root.format = ""
	RootCmd.PersistentPreRunE(listSnapshotsCmd, []string{})
	assert.Equal(t, root.format, snapshotInfoFormat)

	listSnapshots.paging = false
	root.format = ""
	RootCmd.PersistentPreRunE(listSnapshotsCmd, []string{})
	assert.Equal(t, root.format, listSnapshotsFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(createSnapshotCmd, []string{})
	assert.Equal(t, root.format, snapshotInfoFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(createVolumeCmd, []string{})
	assert.Equal(t, root.format, volumeInfoFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(pluginInfoCmd, []string{})
	assert.Equal(t, root.format, pluginInfoFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(pluginCapsCmd, []string{})
	assert.Equal(t, root.format, pluginCapsFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(probeCmd, []string{})
	assert.Equal(t, root.format, probeFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(nodeGetVolumeStatsCmd, []string{})
	assert.Equal(t, root.format, statsFormat)

	root.format = ""
	RootCmd.PersistentPreRunE(nodeGetInfoCmd, []string{})
	assert.Equal(t, root.format, nodeInfoFormat)
}

func TestSettingLogLevel(t *testing.T) {
	// setting debug mode should result in debug level logging
	debug = true
	RootCmd.PersistentPreRunE(nodeGetVolumeStatsCmd, []string{})
	assert.Equal(t, root.logLevel.String(), log.DebugLevel.String())
	// revert back to default
	debug = false
}
