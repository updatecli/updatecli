package cmd

import (
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/updatecli/updatecli/pkg/core/version"
)

var (
	// Version Contains application version
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print current application version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Infof("\n%s\n", "Version")
			version.Show()
		},
	}
)
