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
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// vcs is used to simulate a slice of volume capabilities to test behavior of node commands
var vcs = volumeCapabilitySliceArg{data: []*csi.VolumeCapability{
	{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
	},
}}

func TestNodeCmd(t *testing.T) {
	setupRoot(t, pluginCapsFormat)
	err := nodeCmd.PersistentPreRunE(nodeCmd, []string{})
	assert.NoError(t, err)

	// save original func so we can revert
	cmd := RootCmd.PersistentPreRunE
	// test case: error
	// force RootCmd to return error
	RootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return fmt.Errorf("test error")
	}
	err = nodeCmd.PersistentPreRunE(nodeCmd, []string{})
	assert.ErrorContains(t, err, "test error")

	// restore original func back so other UT won't fail
	RootCmd.PersistentPreRunE = cmd
}

func TestNodeExpandVolumeCmd(t *testing.T) {
	node.client = service.NewClient()
	setupRoot(t, pluginCapsFormat)
	child := nodeExpandVolumeCmd
	err := child.RunE(RootCmd, []string{"volID", "/test/volume"})
	assert.NoError(t, err)

	// try cmd with volume capabilities
	nodeExpandVolume.volCap = vcs
	expandVolume.volCap = vcs

	err = child.RunE(RootCmd, []string{"volID", "/test/volume"})
	assert.NoError(t, err)

	// set req and limit bytes to 1
	nodeExpandVolume.reqBytes = 1
	nodeExpandVolume.limBytes = 1
	err = child.RunE(RootCmd, []string{"volID", "/test/volume"})
	assert.NoError(t, err)

	// force NodeExpandVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"volID", "/test/volume"})
	assert.ErrorContains(t, err, "error from mock NodeExpandVolume")
}

func TestNodeGetCapabilitiesCmd(t *testing.T) {
	node.client = service.NewClient()
	setupRoot(t, pluginCapsFormat)
	child := nodeGetCapabilitiesCmd
	err := child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// force NodeGetCapabilities to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "error from mock NodeGetCapabilities")
}

func TestNodeGetVolumeStatsCmd(t *testing.T) {
	// Set format for NodeGetVolumeStats cmd
	setupRoot(t, statsFormat)
	// root.format = statsFormat

	node.client = service.NewClient()
	child := nodeGetVolumeStatsCmd
	err := child.RunE(RootCmd, []string{"Mock Volume 2:/root/mock-vol:/root/mock/patch"})
	assert.NoError(t, err)

	// force NodeGetVolumeStats to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"Mock Volume 2:/root/mock-vol:/root/mock/patch"})
	assert.ErrorContains(t, err, "error from mock NodeGetVolumeStats")

	// set wrong format to get tpl error, with paging enabled
	setupRoot(t, nodeInfoFormat)
	err = child.RunE(RootCmd, []string{"Mock Volume 2:/root/mock-vol:/root/mock/patch"})
	assert.ErrorContains(t, err, "can't evaluate field NodeId")
}

func TestNodeGetInfo(t *testing.T) {
	// Set format for NodeGetInfo cmd
	setupRoot(t, nodeInfoFormat)

	node.client = service.NewClient()
	child := nodeGetInfoCmd
	err := child.RunE(RootCmd, []string{"mock-node-id"})
	assert.NoError(t, err)

	// force NodeGetInfo to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"mock-node-id"})
	assert.ErrorContains(t, err, "error from mock NodeGetInfo")
}

func TestNodePublishVolume(t *testing.T) {
	setupRoot(t, pluginCapsFormat)
	node.client = service.NewClient()
	child := nodePublishVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.NoError(t, err)

	// try cmd with volume capabilities
	nodePublishVolume.caps = vcs
	err = child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.NoError(t, err)

	// force NodePublishVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.ErrorContains(t, err, "error from mock NodePublishVolume")
}

func TestNodeStageVolume(t *testing.T) {
	setupRoot(t, pluginCapsFormat)
	node.client = service.NewClient()
	child := nodeStageVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.NoError(t, err)

	// try cmd with volume capabilities
	nodeStageVolume.caps = vcs
	err = child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.NoError(t, err)

	// force NodeStageVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"mock-vol-id"})
	assert.ErrorContains(t, err, "error from mock NodeStageVolume")
}

func TestNodeUnpublishVolume(t *testing.T) {
	setupRoot(t, pluginCapsFormat)
	node.client = service.NewClient()
	child := nodeUnpublishVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id", "mock/target/path"})
	assert.NoError(t, err)

	// force NodeUnpublishVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"mock-vol-id", "mock/target/path"})
	assert.ErrorContains(t, err, "error from mock NodeUnpublishVolume")
}

func TestNodeUnstageVolume(t *testing.T) {
	setupRoot(t, pluginCapsFormat)
	node.client = service.NewClient()
	child := nodeUnstageVolumeCmd
	err := child.RunE(RootCmd, []string{"mock-vol-id", "mock/target/path"})
	assert.NoError(t, err)

	// force NodeUnstageVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"mock-vol-id", "mock/target/path"})
	assert.ErrorContains(t, err, "error from mock NodeUnstageVolume")
}
