package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"

	"github.com/spf13/cobra"
)

var (
	pipelineDiffCmd = &cobra.Command{
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
		Use:   "diff NAME[:TAG|@DIGEST]",
		Short: "diff shows changes",
		Run: func(cmd *cobra.Command, args []string) {
			policyReferences = args
			err := getPolicyFilesFromRegistry()
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}

			e.Options.Manifests = append(e.Options.Manifests, manifest.Manifest{
				Manifests: manifestFiles,
				Values:    valuesFiles,
				Secrets:   secretsFiles,
			})

			e.Options.Pipeline.Target.Commit = false
			e.Options.Pipeline.Target.Push = false
			e.Options.Pipeline.Target.Clean = diffClean
			e.Options.Pipeline.Target.DryRun = true

			err = run("pipeline/diff")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	pipelineDiffCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	pipelineDiffCmd.Flags().StringVar(&udashOAuthAudience, "reportAPI", "", "Set the report API URL where to publish pipeline reports")
	pipelineDiffCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	pipelineDiffCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	pipelineDiffCmd.Flags().BoolVar(&diffClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	pipelineDiffCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	pipelineDiffCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their pipeline IDs, accepted a comma separated list")
	pipelineDiffCmd.Flags().StringArrayVar(&labels, "labels", []string{}, "Filter pipelines to apply by their labels, accept a comma separated list (key:value)")

	pipelineCmd.AddCommand(pipelineDiffCmd)
}
