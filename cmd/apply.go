package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"

	"github.com/spf13/cobra"
)

var (
	applyCommit           bool
	applyClean            bool
	applyPush             bool
	applyCleanGitBranches bool
	applyExistingOnly     bool

	applyCmd = &cobra.Command{
		/*
			Technically speaking we could have multiple policies to apply from the command line
			but for clarity we will only allow one policy at a time.
			This decision could be changed in the future.

			The same rule would also apply to the 'diff' and show subcommand.
		*/
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
		Use:   "apply NAME[:TAG|@DIGEST]",
		Short: "apply checks if an update is needed then apply the changes",
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

			e.Options.Pipeline.Target.Commit = applyCommit
			e.Options.Pipeline.Target.Push = applyPush
			e.Options.Pipeline.Target.Clean = applyClean
			e.Options.Pipeline.Target.DryRun = false
			e.Options.Pipeline.Target.CleanGitBranches = applyCleanGitBranches
			e.Options.Pipeline.Target.ExistingOnly = applyExistingOnly

			err = run("apply")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	applyCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	applyCmd.Flags().StringVar(&udashOAuthAudience, "reportAPI", "", "Set the report API URL where to publish pipeline reports")
	applyCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	applyCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	applyCmd.Flags().BoolVar(&applyExistingOnly, "existing-only", false, "Skip targets when pipeline has no existing remote branch '--existing-only=true'")
	applyCmd.Flags().BoolVarP(&applyCommit, "commit", "", true, "Record changes to the repository, '--commit=false'")
	applyCmd.Flags().BoolVarP(&applyPush, "push", "", true, "Update remote refs '--push=false'")
	applyCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	applyCmd.Flags().BoolVar(&applyClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	applyCmd.Flags().BoolVar(&applyCleanGitBranches, "clean-git-branches", false, "Remove updatecli working git branches like '--clean-git-branches=true'")
	applyCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their IDs")
}
