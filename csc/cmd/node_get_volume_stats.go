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
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeGetVolumeStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: `invokes the rpc "NodeGetVolumeStats"`,
	Example: `
USAGE

	csc node stats VOLUME_ID:VOLUME_PATH:STAGING_PATH [VOLUME_ID:VOLUME_PATH:STAGING_PATH...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.NodeGetVolumeStatsRequest{}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID and volume path for the current request.
			split := strings.Split(args[i], ":")
			req.VolumeId, req.VolumePath = split[0], split[1]
			if len(split) > 2 {
				req.StagingTargetPath = split[2]
			}

			log.WithField("request", req).Debug("staging volume")
			rep, err := node.client.NodeGetVolumeStats(ctx, &req)
			if err != nil {
				return err
			}
			if err := root.tpl.Execute(os.Stdout, struct {
				Name string
				Path string
				Resp *csi.NodeGetVolumeStatsResponse
			}{
				Name: req.VolumeId,
				Path: req.VolumePath,
				Resp: rep,
			}); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeGetVolumeStatsCmd)
}
