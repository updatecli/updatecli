package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
)

var (
	manifestShowClean             bool
	manifestShowDisablePrepare    bool
	manifestShowDisableTemplating bool
	manifestShowGraph             bool
	manifestShowGraphFlavour      string

	manifestShowCmd = &cobra.Command{
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
		Use:   "show NAME[:TAG|@DIGEST]",
		Short: "show manifest(s) which will be executed",
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

			e.Options.Pipeline.Target.Clean = manifestShowClean
			e.Options.Config.DisableTemplating = manifestShowDisableTemplating

			if manifestShowGraph {
				// TODO: To be removed once not experimental anymore
				if !experimental {
					logrus.Warningf("The '--graph' flag requires the flag experimental to work.")
					os.Exit(1)
				}
				e.Options.DisplayFlavour = "graph"
				if manifestShowGraphFlavour != "dot" && manifestShowGraphFlavour != "mermaid" {
					logrus.Warningf("The '--graph-flavour' flag should be `dot` or `mermaid`, defaulting to `dot`.")
					manifestShowGraphFlavour = "dot"
				}
				e.Options.GraphFlavour = manifestShowGraphFlavour
			}

			// Showing templating diff may leak sensitive information such as credentials
			config.GolangTemplatingDiff = true

			err = run("manifest/show")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestShowCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	manifestShowCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	manifestShowCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets secrets file uses for templating")
	manifestShowCmd.Flags().BoolVar(&manifestShowClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	manifestShowCmd.Flags().BoolVar(&manifestShowDisablePrepare, "disable-prepare", false, "--disable-prepare skip the Updatecli 'prepare' stage")
	manifestShowCmd.Flags().BoolVar(&manifestShowDisableTemplating, "disable-templating", false, "Disable manifest templating")
	manifestShowCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	manifestShowCmd.Flags().BoolVar(&manifestShowGraph, "graph", false, "Output in graph format")
	manifestShowCmd.Flags().StringVar(&manifestShowGraphFlavour, "graph-flavour", "dot", "Flavour of graph format")

	manifestCmd.AddCommand(manifestShowCmd)
}
