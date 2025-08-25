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

var expandVolume struct {
	reqBytes int64
	limBytes int64
	volCap   volumeCapabilitySliceArg
}

var expandVolumeCmd = &cobra.Command{
	Use:     "expand-volume",
	Aliases: []string{"exp", "expand"},
	Short:   `invokes the rpc "ControllerExpandVolume"`,
	Example: `
USAGE

    csc controller expand [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.ControllerExpandVolumeRequest{
			Secrets: root.secrets,
		}

		if len(expandVolume.volCap.data) > 0 {
			req.VolumeCapability = expandVolume.volCap.data[0]
		}

		if expandVolume.reqBytes > 0 || expandVolume.limBytes > 0 {
			req.CapacityRange = &csi.CapacityRange{}
			if v := expandVolume.reqBytes; v > 0 {
				req.CapacityRange.RequiredBytes = v
			}
			if v := expandVolume.limBytes; v > 0 {
				req.CapacityRange.LimitBytes = v
			}
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("expanding volume")
			rep, err := controller.client.ControllerExpandVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Println(rep.CapacityBytes)
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(expandVolumeCmd)

	flagRequiredBytes(expandVolumeCmd.Flags(), &expandVolume.reqBytes)

	flagLimitBytes(expandVolumeCmd.Flags(), &expandVolume.limBytes)

	flagVolumeCapability(expandVolumeCmd.Flags(), &expandVolume.volCap)

	flagWithRequiresCreds(
		expandVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
