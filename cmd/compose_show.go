package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/compose"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
)

var (
	composeShowCmd = &cobra.Command{
		Use:   "show",
		Short: "show manifest(s) defined by the compose file that should be executed",
		Run: func(cmd *cobra.Command, args []string) {

			composeFiles, err := compose.New(composeCmdFile)
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			manifests := []manifest.Manifest{}
			for i := range composeFiles {
				c := composeFiles[i]
				policies, err := c.GetPolicies(disableTLS)
				if err != nil {
					logrus.Errorf("command failed: %s", err)
					os.Exit(1)
				}

				manifests = append(manifests, policies...)
			}

			e.Options.Manifests = append(e.Options.Manifests, manifests...)

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
	composeShowCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their pipeline IDs, accepted a comma separated list")
	composeShowCmd.Flags().StringArrayVar(&labels, "labels", []string{}, "Filter pipelines to apply by their labels, accepted as a comma separated list (key:value)")

	composeCmd.AddCommand(composeShowCmd)
}
