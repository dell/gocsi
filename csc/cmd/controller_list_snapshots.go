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

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"

	utils "github.com/dell/gocsi/utils/csi"
)

var listSnapshots struct {
	maxEntries     int32
	startingToken  string
	sourceVolumeID string
	SnapshotID     string
	paging         bool
}

var listSnapshotsCmd = &cobra.Command{
	Use:     "list-snapshots",
	Aliases: []string{"sl", "snap-list", "snapshots"},
	Short:   `invokes the rpc "ListSnapshots"`,
	RunE: func(*cobra.Command, []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		req := csi.ListSnapshotsRequest{
			MaxEntries:     listSnapshots.maxEntries,
			StartingToken:  listSnapshots.startingToken,
			SnapshotId:     listSnapshots.SnapshotID,
			SourceVolumeId: listSnapshots.sourceVolumeID,
			Secrets:        root.secrets,
		}

		// If auto-paging is not enabled then send a normal request.
		if !listSnapshots.paging {
			rep, err := controller.client.ListSnapshots(ctx, &req)
			if err != nil {
				return err
			}
			return root.tpl.Execute(os.Stdout, rep)
		}

		// Paging is enabled.
		cvol, cerr := utils.PageSnapshots(ctx, controller.client, req)
		for {
			select {
			case v, ok := <-cvol:
				if !ok {
					return nil
				}
				if err := root.tpl.Execute(os.Stdout, v); err != nil {
					return err
				}
			case e, ok := <-cerr:
				if !ok {
					return nil
				}
				return e
			}
		}
	},
}

func init() {
	controllerCmd.AddCommand(listSnapshotsCmd)

	listSnapshotsCmd.Flags().Int32Var(
		&listSnapshots.maxEntries,
		"max-entries",
		0,
		"The maximum number of entries to return")

	listSnapshotsCmd.Flags().StringVar(
		&listSnapshots.startingToken,
		"starting-token",
		"",
		"The starting token used to retrieve paged data")

	listSnapshotsCmd.Flags().BoolVar(
		&listSnapshots.paging,
		"paging",
		false,
		"Enables auto-paging")
	listSnapshotsCmd.Flags().StringVar(
		&listSnapshots.sourceVolumeID,
		"source-volume-id",
		"",
		"ID of volume to list snapshots for")

	listSnapshotsCmd.Flags().StringVar(
		&listSnapshots.SnapshotID,
		"snapshot-id",
		"",
		"ID of snapshot to retrieve specific snapshot")
	listSnapshotsCmd.Flags().StringVar(
		&root.format,
		"format",
		"",
		"The Go template format used to emit the results")
}
