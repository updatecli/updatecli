package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/updatecli/updatecli/pkg/core/engine"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	valuesFiles  []string
	secretsFiles []string
	e            engine.Engine
	verbose      bool

	rootCmd = &cobra.Command{
		Use:   "updatecli",
		Short: "Updatecli is a tool used to define and apply file update strategies. ",
		Long: `
updatecli is a tool uses to apply file update strategies.
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
		logrus.Errorf("%s %s", result.FAILURE, err)
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
		manifestCmd,
		showCmd,
		versionCmd,
		docsCmd,
		jsonschemaCmd)
}

func run(command string) error {

	switch command {
	case "apply":
		if applyClean {
			defer func() {
				if err := e.Clean(); err != nil {
					logrus.Errorf("error in apply clean - %s", err)
				}
			}()
		}

		err := e.Prepare()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

		err = e.Run()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}
	case "diff":
		if diffClean {
			defer func() {
				if err := e.Clean(); err != nil {
					logrus.Errorf("error in diff clean - %s", err)
				}
			}()
		}

		err := e.Prepare()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

		err = e.Run()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
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

		err := e.Prepare()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
		}

	case "manifest/upgrade":
		err := e.ManifestUpgrade()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	case "show":
		err := e.Show()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	case "jsonschema":
		err := engine.GenerateSchema(jsonschemaBaseID, jsonschemaDirectory)
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}
	default:
		logrus.Warnf("Wrong command")
	}
	return nil
}
