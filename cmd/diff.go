package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"

	"github.com/spf13/cobra"
)

var (
	diffClean bool

	diffCmd = &cobra.Command{
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

			err = run("diff")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	diffCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	diffCmd.Flags().StringVar(&udashOAuthAudience, "reportAPI", "", "Set the report API URL where to publish pipeline reports")
	diffCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	diffCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	diffCmd.Flags().BoolVar(&diffClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	diffCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	diffCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their IDs")
}
