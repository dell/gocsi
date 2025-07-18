package cmd

import (
	"context"
	"fmt"
	"testing"
	"text/template"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	utils "github.com/dell/gocsi/utils/csi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupRoot(t *testing.T, format string) {
	root.ctx = context.Background()
	root.format = format
	tpl, err := template.New("t").Funcs(template.FuncMap{
		"isa": func(o interface{}, t string) bool {
			return fmt.Sprintf("%T", o) == t
		},
	}).Parse(root.format)
	assert.NoError(t, err)
	root.tpl = tpl
}

// By setting a key in the context, we can tell our mock service to return an error
func setupRootCtxToFailCSICalls() {
	returnError := service.ContextKey("returnError")
	root.ctx = context.WithValue(root.ctx, returnError, "true")
}

func TestControllerCmd(t *testing.T) {
	child := controllerCmd

	// test case: no error
	err := child.PersistentPreRunE(child, []string{})
	assert.NoError(t, err)

	// save original func so we can revert
	cmd := RootCmd.PersistentPreRunE

	// test case: error
	// force RootCmd to return error
	RootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return fmt.Errorf("test error")
	}
	err = child.PersistentPreRunE(child, []string{})
	assert.Error(t, err)

	// restore original func back so other UT won't fail
	RootCmd.PersistentPreRunE = cmd
}

func TestCreateSnapshotCmd(t *testing.T) {
	child := createSnapshotCmd
	// set up root as required
	setupRoot(t, snapshotInfoFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	createSnapshot.sourceVol = "Mock Volume 1"
	err := child.RunE(RootCmd, []string{"testname"})
	assert.NoError(t, err)

	// error test case - empty sourceVol
	createSnapshot.sourceVol = ""
	err = child.RunE(RootCmd, []string{"testname"})
	assert.Error(t, err)

	// force CreateSnapshot to return error
	setupRootCtxToFailCSICalls()
	createSnapshot.sourceVol = "Mock Volume 1"
	err = child.RunE(RootCmd, []string{"testname"})
	assert.ErrorContains(t, err, "error from mock CreateSnapshot")

	// set wrong format to get tpl error
	setupRoot(t, nodeInfoFormat)
	err = child.RunE(RootCmd, []string{"testname"})
	assert.ErrorContains(t, err, "can't evaluate field NodeId")
}

func TestCreateVolumeCmd(t *testing.T) {
	child := createVolumeCmd
	// set up root as required
	setupRoot(t, volumeInfoFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	createVolume.reqBytes = 100
	createVolume.limBytes = 200
	createVolume.caps = volumeCapabilitySliceArg{data: []*csi.VolumeCapability{{AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}}
	createVolume.params = mapOfStringArg{data: map[string]string{"key1": "value1", "key2": "value2"}}
	createVolume.sourceVol = "source-volume"
	createVolume.sourceSnap = ""
	err := child.RunE(RootCmd, []string{"testname"})
	assert.NoError(t, err)

	// Valid test case 2: snapshot
	createVolume.sourceVol = ""
	createVolume.sourceSnap = "source-snap"
	err = child.RunE(RootCmd, []string{"testname"})
	assert.NoError(t, err)

	// Error test case: have both source vol and source snap
	createVolume.sourceVol = "source-volume"
	err = child.RunE(RootCmd, []string{"testname"})
	assert.Error(t, err)

	// force CreateVolume to return error
	createVolume.sourceVol = ""
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"testname"})
	assert.ErrorContains(t, err, "error from mock CreateVolume")

	// set wrong format to get tpl error
	setupRoot(t, nodeInfoFormat)
	err = child.RunE(RootCmd, []string{"testname"})
	assert.ErrorContains(t, err, "can't evaluate field NodeId")
}

func TestDeleteSnapshotCmd(t *testing.T) {
	child := deleteSnapshotCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	err := child.RunE(RootCmd, []string{"testname"})
	assert.NoError(t, err)

	// force DeleteSnapshot to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"testname"})
	assert.ErrorContains(t, err, "error from mock DeleteSnapshot")
}

func TestDeleteVolumeCmd(t *testing.T) {
	child := deleteVolumeCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	err := child.RunE(RootCmd, []string{"testname"})
	assert.NoError(t, err)

	// force DeleteVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"testname"})
	assert.ErrorContains(t, err, "error from mock DeleteVolume")
}

func TestExpandVolumeCmd(t *testing.T) {
	child := expandVolumeCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	expandVolume.reqBytes = 2 * utils.Gib100
	expandVolume.limBytes = 3 * utils.Gib100
	expandVolume.volCap = volumeCapabilitySliceArg{data: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}}}}
	err := child.RunE(RootCmd, []string{"1"}) // uses volume ID, which starts at 1 with our mocks
	assert.NoError(t, err)

	// force ExpandVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"1"})
	assert.ErrorContains(t, err, "error from mock ControllerExpandVolume")
}

func TestGetCapabilitiesCmd(t *testing.T) {
	child := controllerGetCapabilitiesCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	err := child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// force GetCapabilities to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "error from mock ControllerGetCapabilities")
}

func TestGetCapacityCmd(t *testing.T) {
	child := getCapacityCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	getCapacity.caps.data = []*csi.VolumeCapability{
		{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
	}
	getCapacity.params.data = map[string]string{}
	err := child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// force GetCapacity to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "error from mock GetCapacity")
}

func TestListSnapshotsCmd(t *testing.T) {
	child := listSnapshotsCmd
	// set up root as required
	setupRoot(t, snapshotInfoFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	listSnapshots.maxEntries = 10
	listSnapshots.startingToken = "1"
	listSnapshots.sourceVolumeID = "1"
	listSnapshots.SnapshotID = "1"
	listSnapshots.paging = true
	err := child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// do it again, but with paging disabled
	listSnapshots.paging = false
	setupRoot(t, listSnapshotsFormat)
	err = child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// force ListSnapshots to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "error from mock ListSnapshots")

	// set wrong format to get tpl error, with paging enabled
	listSnapshots.paging = true
	setupRoot(t, nodeInfoFormat)
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "can't evaluate field NodeId")

	// TODO: more error cases
}

func TestListVolumesCmd(t *testing.T) {
	child := listVolumesCmd
	// set up root as required
	setupRoot(t, volumeInfoFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	listVolumes.maxEntries = 10
	listVolumes.startingToken = "1"
	listVolumes.paging = true
	err := child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// do it again, but with paging disabled
	listVolumes.paging = false
	setupRoot(t, listVolumesFormat)
	err = child.RunE(RootCmd, []string{})
	assert.NoError(t, err)

	// force ListVolumes to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "error from mock ListVolumes")

	// set wrong format to get tpl error, with paging enabled
	listVolumes.paging = true
	setupRoot(t, nodeInfoFormat)
	err = child.RunE(RootCmd, []string{})
	assert.ErrorContains(t, err, "can't evaluate field NodeId")
	// TODO: error cases
}

func TestPublishVolumeCmd(t *testing.T) {
	child := controllerPublishVolumeCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	controllerPublishVolume.nodeID = "node1"
	controllerPublishVolume.caps.data = []*csi.VolumeCapability{
		{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
	}
	controllerPublishVolume.volCtx.data = map[string]string{}
	controllerPublishVolume.readOnly = false
	err := child.RunE(RootCmd, []string{"1"})
	assert.NoError(t, err)

	// force ControllerPublishVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"1"})
	assert.ErrorContains(t, err, "error from mock ControllerPublishVolume")
}

func TestUnpublishVolumeCmd(t *testing.T) {
	child := controllerUnpublishVolumeCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	controllerUnpublishVolume.nodeID = "node1"
	err := child.RunE(RootCmd, []string{"1"})
	assert.NoError(t, err)

	// force ControllerUnpublishVolume to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"1"})
	assert.ErrorContains(t, err, "error from mock ControllerUnpublishVolume")
}

func TestValidateVolumeCapabilitiesCmd(t *testing.T) {
	child := valVolCapsCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	valVolCaps.volCtx.data = map[string]string{}
	valVolCaps.params.data = map[string]string{}
	valVolCaps.caps.data = []*csi.VolumeCapability{
		{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
	}
	err := child.RunE(RootCmd, []string{"1"})
	assert.NoError(t, err)

	// force ValidateVolumeCapabilities to return error
	setupRootCtxToFailCSICalls()
	err = child.RunE(RootCmd, []string{"1"})
	assert.ErrorContains(t, err, "error from mock ValidateVolumeCapabilities")
}
