package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"

	"github.com/spf13/cobra"
)

var (
	prepareClean bool

	prepareCmd = &cobra.Command{
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
		Use:   "prepare NAME[:TAG|@DIGEST]",
		Short: "prepare run tasks needed for a run like `git clone`",
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

			e.Options.Pipeline.Target.Clean = prepareClean

			logrus.Warningln("Deprecated command, please instead use `updatecli pipeline prepare`")

			err = run("prepare")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	prepareCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	prepareCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	prepareCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	prepareCmd.Flags().BoolVar(&prepareClean, "clean", false, "Remove updatecli working directory like '--clean=true")
	prepareCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	prepareCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their pipeline IDs, accepted a comma separated list")
	prepareCmd.Flags().StringArrayVar(&labels, "labels", []string{}, "Filter pipelines to apply by their labels, accepted as a comma separated list (key:value)")
}
