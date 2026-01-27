package cmd

import (
	"github.com/spf13/cobra"
)

var (
	pipelineCmd = &cobra.Command{
		Use:   "pipeline",
		Short: "pipeline executes specific pipeline tasks such as diff or apply",
	}
)
