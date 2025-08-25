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

var nodePublishVolume struct {
	targetPath        string
	stagingTargetPath string
	pubCtx            mapOfStringArg
	volCtx            mapOfStringArg
	attribs           mapOfStringArg
	readOnly          bool
	caps              volumeCapabilitySliceArg
}

var nodePublishVolumeCmd = &cobra.Command{
	Use:     "publish",
	Aliases: []string{"mnt", "mount"},
	Short:   `invokes the rpc "NodePublishVolume"`,
	Example: `
USAGE

    csc node publish [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.NodePublishVolumeRequest{
			StagingTargetPath: nodePublishVolume.stagingTargetPath,
			TargetPath:        nodePublishVolume.targetPath,
			PublishContext:    nodePublishVolume.pubCtx.data,
			Readonly:          nodePublishVolume.readOnly,
			Secrets:           root.secrets,
			VolumeContext:     nodePublishVolume.volCtx.data,
		}

		if len(nodePublishVolume.caps.data) > 0 {
			req.VolumeCapability = nodePublishVolume.caps.data[0]
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("mounting volume")
			_, err := node.client.NodePublishVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodePublishVolumeCmd)

	flagStagingTargetPath(
		nodePublishVolumeCmd.Flags(), &nodePublishVolume.stagingTargetPath)

	flagTargetPath(
		nodePublishVolumeCmd.Flags(), &nodePublishVolume.targetPath)

	flagReadOnly(
		nodePublishVolumeCmd.Flags(), &nodePublishVolume.readOnly)

	flagVolumeContext(nodePublishVolumeCmd.Flags(), &nodePublishVolume.volCtx)

	flagPublishContext(nodePublishVolumeCmd.Flags(), &nodePublishVolume.pubCtx)

	flagVolumeCapability(
		nodePublishVolumeCmd.Flags(), &nodePublishVolume.caps)

	flagWithRequiresVolContext(
		nodePublishVolumeCmd.Flags(), &root.withRequiresVolContext, false)

	flagWithRequiresPubContext(
		nodePublishVolumeCmd.Flags(), &root.withRequiresPubContext, false)

	flagWithRequiresCreds(
		nodePublishVolumeCmd.Flags(), &root.withRequiresCreds, "")
}
