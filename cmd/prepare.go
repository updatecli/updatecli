package cmd

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	prepareClean bool

	prepareCmd = &cobra.Command{
		Use:   "prepare",
		Short: "prepare run tasks needed for a run like `git clone`",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Infof("\n%s\n", strings.ToTitle("Prepare"))

			e.Options.File = cfgFile
			e.Options.ValuesFile = valuesFile

			e.Options.Target.Clean = prepareClean

			err := run("prepare")
			if err != nil {
				logrus.Errorf("command failed")
				os.Exit(1)
			}
		},
	}
)

func init() {
	prepareCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "Sets config file or directory. (default: './updateCli.yaml')")
	prepareCmd.Flags().StringVarP(&valuesFile, "values", "v", "", "Sets values file uses for templating (required {.tpl,.tmpl} config)")
	prepareCmd.Flags().BoolVarP(&prepareClean, "clean", "", false, "Remove updatecli working directory like '--clean=true '(default: false)")
}
