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

var controllerUnpublishVolume struct {
	nodeID string
}

var controllerUnpublishVolumeCmd = &cobra.Command{
	Use:     "unpublish",
	Aliases: []string{"detach"},
	Short:   `invokes the rpc "ControllerUnpublishVolume"`,
	Example: `
USAGE

    csc controller unpublishvolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.ControllerUnpublishVolumeRequest{
			NodeId:  controllerUnpublishVolume.nodeID,
			Secrets: root.secrets,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("unpublishing volume")
			_, err := controller.client.ControllerUnpublishVolume(ctx, &req)
			if err != nil {
				return err
			}
			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(controllerUnpublishVolumeCmd)

	controllerUnpublishVolumeCmd.Flags().StringVar(
		&controllerUnpublishVolume.nodeID,
		"node-id",
		"",
		"The ID of the node from which to unpublish the volume")

	flagWithRequiresCreds(
		controllerUnpublishVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
