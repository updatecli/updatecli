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
	composeShowOnlyPolicyIDs    []string
	composeShowIgnoredPolicyIDs []string

	composeShowCmd = &cobra.Command{
		Use:   "show",
		Short: "show manifest(s) defined by the compose file that should be executed",
		Run: func(cmd *cobra.Command, args []string) {

			composeFiles, err := compose.New(composeCmdFile, map[string]bool{})
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			manifests := []manifest.Manifest{}
			for i := range composeFiles {
				c := composeFiles[i]
				policies, err := c.GetPolicies(
					disableTLS,
					parseParametersList(composeShowOnlyPolicyIDs),
					parseParametersList(composeShowIgnoredPolicyIDs),
				)
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
	composeShowCmd.Flags().StringArrayVar(&composeShowOnlyPolicyIDs, "only-policy-ids", []string{}, "Filter policies to apply by their policy IDs, accepted as a comma separated list")
	composeShowCmd.Flags().StringArrayVar(&composeShowIgnoredPolicyIDs, "ignored-policy-ids", []string{}, "Filter policies to ignore by their policy IDs, accepted as a comma separated list")

	composeCmd.AddCommand(composeShowCmd)
}
