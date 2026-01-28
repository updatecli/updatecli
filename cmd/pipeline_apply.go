package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"

	"github.com/spf13/cobra"
)

var (
	pipelineApplyCmd = &cobra.Command{
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

			err = run("pipeline/apply")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	pipelineApplyCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	pipelineApplyCmd.Flags().StringVar(&udashOAuthAudience, "reportAPI", "", "Set the report API URL where to publish pipeline reports")
	pipelineApplyCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	pipelineApplyCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets Sops secrets file uses for templating")
	pipelineApplyCmd.Flags().BoolVar(&applyExistingOnly, "existing-only", false, "Skip targets when pipeline has no existing remote branch '--existing-only=true'")
	pipelineApplyCmd.Flags().BoolVarP(&applyCommit, "commit", "", true, "Record changes to the repository, '--commit=false'")
	pipelineApplyCmd.Flags().BoolVarP(&applyPush, "push", "", true, "Update remote refs '--push=false'")
	pipelineApplyCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")
	pipelineApplyCmd.Flags().BoolVar(&applyClean, "clean", false, "Remove updatecli working directory like '--clean=true'")
	pipelineApplyCmd.Flags().BoolVar(&applyCleanGitBranches, "clean-git-branches", false, "Remove updatecli working git branches like '--clean-git-branches=true'")
	pipelineApplyCmd.Flags().StringArrayVar(&pipelineIds, "pipeline-ids", []string{}, "Filter pipelines to apply by their IDs")
	pipelineApplyCmd.Flags().StringArrayVar(&labels, "labels", []string{}, "Filter pipelines to apply by their labels, accept a comma separated list (key:value)")

	pipelineCmd.AddCommand(pipelineApplyCmd)
}
