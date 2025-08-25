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

var deleteSnapshotCmd = &cobra.Command{
	Use:     "delete-snapshot",
	Aliases: []string{"ds", "delsnap"},
	Short:   `invokes the rpc "DeleteSnapshot"`,
	Example: `
USAGE

    csc controller delete-snapshot [flags] snapshot_ID [snapshot_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.DeleteSnapshotRequest{
			Secrets: root.secrets,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the snapshot ID for the current request.
			req.SnapshotId = args[i]

			log.WithField("request", req).Debug("deleting snapshot")
			_, err := controller.client.DeleteSnapshot(ctx, &req)
			if err != nil {
				return err
			}
			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(deleteSnapshotCmd)

	flagWithRequiresCreds(
		deleteSnapshotCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
