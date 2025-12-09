package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/compose"
	"github.com/updatecli/updatecli/pkg/core/config"
)

var (
	composeShowCmd = &cobra.Command{
		Use:   "show",
		Short: "show manifest(s) defined by the compose file that should be executed",
		Run: func(cmd *cobra.Command, args []string) {

			c, err := compose.New(composeCmdFile)
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			policies, err := c.GetPolicies(disableTLS)
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			e.Options.Manifests = append(e.Options.Manifests, policies...)

			e.Options.Pipeline.Target.Clean = composeCmdClean
			e.Options.Config.DisableTemplating = composeCmdDisableTemplating

			// Showing templating diff may leak sensitive information such as credentials
			config.GolangTemplatingDiff = true

			err = run("compose/show")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	composeShowCmd.Flags().StringVarP(&composeCmdFile, "file", "f", composeDefaultCmdFile, "Define the update-compose file")
	composeShowCmd.Flags().BoolVar(&composeCmdClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	composeShowCmd.Flags().BoolVar(&composeCmdDisablePrepare, "disable-prepare", false, "--disable-prepare skip the Updatecli 'prepare' stage")
	composeShowCmd.Flags().BoolVar(&composeCmdDisableTemplating, "disable-templating", false, "Disable manifest templating")
	composeShowCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	composeShowCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their IDs, accepted a commma separated list")

	composeCmd.AddCommand(composeShowCmd)
}
