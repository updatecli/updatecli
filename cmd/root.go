package cmd

import (
	"github.com/olblak/updateCli/pkg/core/log"
	"github.com/sirupsen/logrus"
	"os"

	"github.com/olblak/updateCli/pkg/core/engine"
	"github.com/olblak/updateCli/pkg/core/result"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	valuesFile string
	e engine.Engine
    verbose bool

rootCmd = &cobra.Command{
		Use:   "updateCli",
		Short: "Updatecli is a tool used to define and apply file update strategies. ",
		Long: `
Updatecli is a tool uses to apply file update strategies.
It reads a yaml or a go template configuration file, then works into three stages:

1. Source: Based on a rule fetch a value that will be injected in later stages.
2. Conditions: Ensure that conditions are met based on the value retrieved during the source stage.
3. Target: Update and publish the target files based on a value retrieved from the source stage.
`,
	}
)

// Execute executes the root command.
func Execute() {
	logrus.SetFormatter(log.NewTextFormat())

	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("\n\u26A0 %s \n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "debug", "", false, "Debug Output")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}
	rootCmd.AddCommand(
		applyCmd,
		diffCmd,
		prepareCmd,
		showCmd,
		versionCmd,
		docsCmd)
}

func run(command string) error {

	switch command {
	case "apply":
		err := e.Prepare()
		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
			return err
		}

		if applyClean {
			defer func() {
				if err := e.Clean(); err != nil {
					logrus.Errorf("error in apply clean - %s", err)
				}
			}()
		}

		err = e.Run()
		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
			return err
		}
	case "diff":
		err := e.Prepare()
		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
			return err
		}

		if diffClean {
			defer func() {
				if err := e.Clean(); err != nil {
					logrus.Errorf("error in diff clean - %s", err)
				}
			}()
		}

		err = e.Run()
		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
			return err
		}
	case "prepare":
		if prepareClean {
			defer func() {
				if err := e.Clean(); err != nil {
					logrus.Errorf("error in prepare clean - %s", err)
				}
			}()
		}
	case "show":
		err := e.Show()
		if err != nil {
			logrus.Errorf("\n%s %s \n\n", result.FAILURE, err)
			return err
		}
	default:
		logrus.Warnf("Wrong command")
	}
	return nil
}
