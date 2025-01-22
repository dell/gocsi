package cmd

import (
	"context"
	"fmt"
	"testing"
	"text/template"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestControllerCmd(t *testing.T) {
	child := controllerCmd

	// test case: no error
	err := child.PersistentPreRunE(child, []string{})
	assert.NoError(t, err)

	// save original func so we can revert
	cmd := RootCmd.PersistentPreRunE

	// test case: error
	// force RootCmd to return error
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
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
	root.ctx = context.Background()
	root.timeout = 10 * time.Second
	tpl, err := template.New("t").Funcs(template.FuncMap{
		"isa": func(o interface{}, t string) bool {
			return fmt.Sprintf("%T", o) == t
		},
	}).Parse(root.format)
	assert.NoError(t, err)
	root.tpl = tpl

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	createSnapshot.sourceVol = "Mock Volume 1"
	err = child.RunE(RootCmd, []string{"testname"})
	assert.NoError(t, err)

	// error test case - empty sourceVol
	createSnapshot.sourceVol = ""
	err = child.RunE(RootCmd, []string{"testname"})
	assert.Error(t, err)

	// TODO: CreateSnapshot error case, tpl.Execute() error case
}

func TestCreateVolumeCmd(t *testing.T) {
	child := createVolumeCmd
	// set up root as required
	root.ctx = context.Background()
	root.timeout = 10 * time.Second
	tpl, err := template.New("t").Funcs(template.FuncMap{
		"isa": func(o interface{}, t string) bool {
			return fmt.Sprintf("%T", o) == t
		},
	}).Parse(root.format)
	assert.NoError(t, err)
	root.tpl = tpl

	// set up the CSI client with a mock
	controller.client = service.NewClient()

	// Valid test case
	createVolume.reqBytes = 100
	createVolume.limBytes = 200
	createVolume.caps = volumeCapabilitySliceArg{data: []*csi.VolumeCapability{{AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}}
	createVolume.params = mapOfStringArg{data: map[string]string{"key1": "value1", "key2": "value2"}}
	createVolume.sourceVol = "source-volume"
	createVolume.sourceSnap = ""
	err = child.RunE(RootCmd, []string{"testname"})
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

	// TODO: more error test cases
}
