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

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeGetCapabilitiesCmd = &cobra.Command{
	Use:     "get-capabilities",
	Aliases: []string{"capabilities"},
	Short:   `invokes the rpc "NodeGetCapabilities"`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := node.client.NodeGetCapabilities(
			ctx,
			&csi.NodeGetCapabilitiesRequest{})
		if err != nil {
			return err
		}

		for _, cap := range rep.Capabilities {
			fmt.Println(cap.Type)
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeGetCapabilitiesCmd)
}
