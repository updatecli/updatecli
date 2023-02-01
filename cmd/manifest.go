package cmd

import (
	"github.com/spf13/cobra"
)

var (
	manifestCmd = &cobra.Command{
		Use:   "manifest",
		Short: "manifest executes specific manifest task such as upgrade",
	}
)
