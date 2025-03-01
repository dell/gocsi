package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var pluginInfoCmd = &cobra.Command{
	Use:     "plugin-info",
	Aliases: []string{"info"},
	Short:   `invokes the rpc "GetPluginInfo"`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := identity.client.GetPluginInfo(
			ctx,
			&csi.GetPluginInfoRequest{})
		if err != nil {
			return err
		}

		return root.tpl.Execute(getStdout(), rep)
	},
}

func init() {
	identityCmd.AddCommand(pluginInfoCmd)
}
