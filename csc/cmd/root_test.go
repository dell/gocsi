/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
