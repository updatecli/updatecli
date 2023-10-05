package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	// manifestPushPolicyReference is the OCI registry reference to push
	manifestPushPolicyReference []string
	// manifestPushFileStore is the path to the manifest(s) root directory from where to push
	manifestPushFileStore string
	// manifestPushPolicyFile is the path to the policy file containing policy metadata infomation
	manifestPushPolicyFile string

	// manifestPushCmd is the Cobra command to push OCI registry manifest(s)
	manifestPushCmd = &cobra.Command{
		Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		Use:   "push [PATH]",
		Short: "push manifest(s) to an OCI registry",
		Run: func(cmd *cobra.Command, args []string) {
			manifestPushFileStore = args[0]

			err := run("manifest/push")
			if err != nil {
				logrus.Errorf("command failed: %s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	manifestPushCmd.Flags().StringVar(&manifestPushPolicyFile, "policy", "Policy.yaml", "Sets policy file containing policy metadata infomation")
	manifestPushCmd.Flags().StringArrayVarP(&manifestFiles, "config", "c", []string{}, "Sets config file or directory. By default, Updatecli looks for a file named 'updatecli.yaml' or a directory named 'updatecli.d'")
	manifestPushCmd.Flags().StringArrayVarP(&valuesFiles, "values", "v", []string{}, "Sets values file uses for templating")
	manifestPushCmd.Flags().StringArrayVarP(&manifestPushPolicyReference, "tag", "t", []string{}, `Name and optionally a tag (format: "name:tag")`)
	manifestPushCmd.Flags().StringArrayVar(&secretsFiles, "secrets", []string{}, "Sets secrets file uses for templating")
	manifestPushCmd.Flags().BoolVar(&disableTLS, "disable-tls", false, "Disable TLS verification like '--disable-tls=true'")

	manifestCmd.AddCommand(manifestPushCmd)
}
