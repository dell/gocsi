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

var valVolCaps struct {
	volCtx mapOfStringArg
	params mapOfStringArg
	caps   volumeCapabilitySliceArg
}

var valVolCapsCmd = &cobra.Command{
	Use:     "validate-volume-capabilities",
	Aliases: []string{"validate"},
	Short:   `invokes the rpc "ValidateVolumeCapabilities"`,
	Example: `
USAGE

    csc controller validate [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		req := csi.ValidateVolumeCapabilitiesRequest{
			VolumeContext:      valVolCaps.volCtx.data,
			VolumeCapabilities: valVolCaps.caps.data,
			Parameters:         valVolCaps.params.data,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("validate volume capabilities")
			rep, err := controller.client.ValidateVolumeCapabilities(ctx, &req)
			if err != nil {
				return err
			}
			fmt.Printf("%q\t%v", args[i], rep.Confirmed)
			if rep.Message != "" {
				fmt.Printf("\t%q", rep.Message)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(valVolCapsCmd)

	flagParameters(valVolCapsCmd.Flags(), &valVolCaps.params)

	flagVolumeCapabilities(valVolCapsCmd.Flags(), &valVolCaps.caps)

	flagVolumeContext(valVolCapsCmd.Flags(), &valVolCaps.volCtx)

	flagWithRequiresVolContext(
		valVolCapsCmd.Flags(), &root.withRequiresVolContext, false)
}
