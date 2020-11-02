package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/olblak/updateCli/pkg/core/version"
)

var (
	// Version Contains application version
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print current application version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%s\n\n", strings.ToTitle("Version"))
			version.Show()
		},
	}
)
