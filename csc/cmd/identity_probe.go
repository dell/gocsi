package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: `invokes the rpc "Probe"`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := identity.client.Probe(
			ctx,
			&csi.ProbeRequest{})
		if err != nil {
			return err
		}

		return root.tpl.Execute(getStdout(), rep)
	},
}

func init() {
	identityCmd.AddCommand(probeCmd)
}
