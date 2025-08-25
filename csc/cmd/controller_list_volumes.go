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

var listVolumes struct {
	maxEntries    int32
	startingToken string
	paging        bool
}

var listVolumesCmd = &cobra.Command{
	Use:     "list-volumes",
	Aliases: []string{"ls", "list", "volumes"},
	Short:   `invokes the rpc "ListVolumes"`,
	RunE: func(*cobra.Command, []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		req := csi.ListVolumesRequest{
			MaxEntries:    listVolumes.maxEntries,
			StartingToken: listVolumes.startingToken,
		}

		// If auto-paging is not enabled then send a normal request.
		if !listVolumes.paging {
			rep, err := controller.client.ListVolumes(ctx, &req)
			if err != nil {
				return err
			}
			return root.tpl.Execute(os.Stdout, rep)
		}

		// Paging is enabled.
		cvol, cerr := utils.PageVolumes(ctx, controller.client, req)
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
	controllerCmd.AddCommand(listVolumesCmd)

	listVolumesCmd.Flags().Int32Var(
		&listVolumes.maxEntries,
		"max-entries",
		0,
		"The maximum number of entries to return")

	listVolumesCmd.Flags().StringVar(
		&listVolumes.startingToken,
		"starting-token",
		"",
		"The starting token used to retrieve paged data")

	listVolumesCmd.Flags().BoolVar(
		&listVolumes.paging,
		"paging",
		false,
		"Enables auto-paging")

	listVolumesCmd.Flags().StringVar(
		&root.format,
		"format",
		"",
		"The Go template format used to emit the results")
}
