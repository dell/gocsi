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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeExpandVolume struct {
	reqBytes    int64
	limBytes    int64
	stagingPath string
	volCap      volumeCapabilitySliceArg
}

var nodeExpandVolumeCmd = &cobra.Command{
	Use:     "expand-volume",
	Aliases: []string{"exp", "expand"},
	Short:   `invokes the rpc "NodeExpandVolume"`,
	Example: `
USAGE

    csc node expand [flags] VOLUME_ID VOLUME_PATH
`,
	Args: cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		// Set the volume name and path for the current request.
		req := csi.NodeExpandVolumeRequest{
			VolumeId:          args[0],
			VolumePath:        args[1],
			StagingTargetPath: nodeExpandVolume.stagingPath,
		}

		if len(nodeExpandVolume.volCap.data) > 0 {
			req.VolumeCapability = expandVolume.volCap.data[0]
		}

		if nodeExpandVolume.reqBytes > 0 || nodeExpandVolume.limBytes > 0 {
			req.CapacityRange = &csi.CapacityRange{}
			if v := nodeExpandVolume.reqBytes; v > 0 {
				req.CapacityRange.RequiredBytes = v
			}
			if v := nodeExpandVolume.limBytes; v > 0 {
				req.CapacityRange.LimitBytes = v
			}
		}

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		log.WithField("request", req).Debug("expanding volume")
		rep, err := node.client.NodeExpandVolume(ctx, &req)
		if err != nil {
			return err
		}

		fmt.Println(rep.CapacityBytes)

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeExpandVolumeCmd)

	flagRequiredBytes(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.reqBytes)

	flagLimitBytes(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.limBytes)

	flagStagingTargetPath(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.stagingPath)

	flagVolumeCapability(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.volCap)
}
