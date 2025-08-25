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
	"context"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var createVolume struct {
	reqBytes   int64
	limBytes   int64
	caps       volumeCapabilitySliceArg
	params     mapOfStringArg
	sourceVol  string
	sourceSnap string
}

var createVolumeCmd = &cobra.Command{
	Use:     "create-volume",
	Aliases: []string{"n", "c", "new", "create"},
	Short:   `invokes the rpc "CreateVolume"`,
	Example: `
CREATING MULTIPLE VOLUMES
        The following example illustrates how to create two volumes with the
        same characteristics at the same time:

            csc controller new --endpoint /csi/server.sock
                               --cap 1,block \
                               --cap MULTI_NODE_MULTI_WRITER,mount,xfs,uid=500 \
                               --params region=us,zone=texas
                               --params disabled=false
                               MyNewVolume1 MyNewVolume2
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.CreateVolumeRequest{
			VolumeCapabilities: createVolume.caps.data,
			Parameters:         createVolume.params.data,
			Secrets:            root.secrets,
		}

		if createVolume.reqBytes > 0 || createVolume.limBytes > 0 {
			req.CapacityRange = &csi.CapacityRange{}
			if v := createVolume.reqBytes; v > 0 {
				req.CapacityRange.RequiredBytes = v
			}
			if v := createVolume.limBytes; v > 0 {
				req.CapacityRange.LimitBytes = v
			}
		}

		if createVolume.sourceVol != "" && createVolume.sourceSnap != "" {
			return errors.New(
				"--source-volume and --source-snapshot are mutually exclusive")
		}
		if createVolume.sourceVol != "" {
			req.VolumeContentSource = &csi.VolumeContentSource{
				Type: &csi.VolumeContentSource_Volume{
					Volume: &csi.VolumeContentSource_VolumeSource{
						VolumeId: createVolume.sourceVol,
					},
				},
			}
		} else if createVolume.sourceSnap != "" {
			req.VolumeContentSource = &csi.VolumeContentSource{
				Type: &csi.VolumeContentSource_Snapshot{
					Snapshot: &csi.VolumeContentSource_SnapshotSource{
						SnapshotId: createVolume.sourceSnap,
					},
				},
			}
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.Name = args[i]

			log.WithField("request", req).Debug("creating volume")
			rep, err := controller.client.CreateVolume(ctx, &req)
			if err != nil {
				return err
			}
			if err := root.tpl.Execute(os.Stdout, rep.Volume); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(createVolumeCmd)

	flagRequiredBytes(createVolumeCmd.Flags(), &createVolume.reqBytes)

	flagLimitBytes(createVolumeCmd.Flags(), &createVolume.limBytes)

	flagParameters(createVolumeCmd.Flags(), &createVolume.params)

	flagVolumeCapabilities(createVolumeCmd.Flags(), &createVolume.caps)

	createVolumeCmd.Flags().StringVar(
		&createVolume.sourceVol,
		"source-volume",
		"",
		"Pre-populate data using a source volume")

	createVolumeCmd.Flags().StringVar(
		&createVolume.sourceSnap,
		"source-snapshot",
		"",
		"Pre-populate data using a source snapshot")

	flagWithRequiresVolContext(
		createVolumeCmd.Flags(),
		&root.withRequiresVolContext,
		false)

	flagWithRequiresCreds(
		createVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
