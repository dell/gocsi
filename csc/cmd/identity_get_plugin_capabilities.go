package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var pluginCapsCmd = &cobra.Command{
	Use:     "plugin-capabilities",
	Aliases: []string{"caps"},
	Short:   `invokes the rpc "GetPluginCapabilities"`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := identity.client.GetPluginCapabilities(
			ctx,
			&csi.GetPluginCapabilitiesRequest{})
		if err != nil {
			return err
		}

		return root.tpl.Execute(getStdout(), rep)
	},
}

func init() {
	identityCmd.AddCommand(pluginCapsCmd)
}
