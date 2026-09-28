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

	log "github.com/dell/csmlog"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var modifyVolume struct {
	mutableParams mapOfStringArg
}

var modifyVolumeCmd = &cobra.Command{
	Use:     "modify-volume",
	Aliases: []string{"modify", "mod"},
	Short:   `invokes the rpc "ControllerModifyVolume"`,
	Example: `
USAGE

    csc controller modify-volume --mutable-params key1=val1,key2=val2 VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.ControllerModifyVolumeRequest{
			MutableParameters: modifyVolume.mutableParams.data,
			Secrets:           root.secrets,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)

			req.VolumeId = args[i]

			log.WithFields(log.Fields{"request": &req}).Debug("modifying volume")
			_, err := controller.client.ControllerModifyVolume(ctx, &req)
			cancel()
			if err != nil {
				return err
			}
			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(modifyVolumeCmd)

	modifyVolumeCmd.Flags().Var(
		&modifyVolume.mutableParams,
		"mutable-params",
		`One or more key/value pairs may be specified to send with
        the request as its MutableParameters field:

            --mutable-params key1=val1,key2=val2 --mutable-params=key3=val3`)

	flagWithRequiresCreds(
		modifyVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
