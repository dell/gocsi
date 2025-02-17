package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func Test_argName(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		want string
	}{
		{
			name: "test createVolumeCmd",
			cmd:  createVolumeCmd,
			want: "VOLUME_NAME [VOLUME_NAME...]",
		},
		{
			name: "test deleteVolumeCmd",
			cmd:  deleteVolumeCmd,
			want: "VOLUME_ID [VOLUME_ID...]",
		},
		{
			name: "test rootcmd",
			cmd:  RootCmd,
			want: "CMD",
		},
		{
			name: "test non recognized command",
			cmd:  &cobra.Command{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := argName(tt.cmd); got != tt.want {
				t.Errorf("argName returned %v, expected: %v", got, tt.want)
			}
		})
	}
}

func Test_helpFunc(t *testing.T) {
	// set up test command for testing helpFunc when help flag defaults to false
	testCmd := &cobra.Command{Use: "test-cmd"}
	flags := pflag.NewFlagSet("test-flag-set-name", pflag.ContinueOnError)
	var myFlag bool
	flags.BoolVar(&myFlag, "help", false, "help with test cmd")
	testCmd.Flags().AddFlagSet(flags)

	type args struct {
		cmd *cobra.Command
		in1 []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test help function with deleteVolumeCmd",
			args: args{
				cmd: deleteVolumeCmd,
				in1: []string{},
			},
		},
		{
			name: "test help function with help flag set to false",
			args: args{
				cmd: testCmd,
				in1: []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// helpFunc does not return any errors or values, so verify panic wasn't hit
			assert.NotPanics(t, func() { helpFunc(tt.args.cmd, tt.args.in1) })
		})
	}
}

func Test_usageFunc(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{
			name: "test usage function with deleteVolumeCmd",
			cmd:  deleteVolumeCmd,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := usageFunc(tt.cmd)
			assert.NoError(t, err)
		})
	}
}
