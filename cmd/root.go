package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/registry"
	"github.com/updatecli/updatecli/pkg/core/udash"

	"github.com/updatecli/updatecli/pkg/core/engine"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/spf13/cobra"
)

var (
	manifestFiles    []string
	valuesFiles      []string
	secretsFiles     []string
	policyReferences []string
	e                engine.Engine
	verbose          bool
	experimental     bool
	disableTLS       bool

	rootCmd = &cobra.Command{
		Use:   "updatecli",
		Short: "Updatecli is a declarative dependency manager command line tool",
		Long: `
Updatecli is a declarative dependency manager command line tool.
Based on Updatecli manifest(s), It ensures that target files are up to date.
Updatecli  works into three stages:

1. Source: Retrieve a value from a third location like file, api, etc..
2. Condition: Ensure conditions are met based on the value retrieved during the source stage.
3. Target: Update the target based on the value retrieved from the source stage.
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
	rootCmd.PersistentFlags().BoolVarP(&experimental, "experimental", "", false, "Enable Experimental mode")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if experimental {
			cmdoptions.Experimental = true
			logrus.Infof("Experimental Mode Enabled")
		}
	}
	rootCmd.AddCommand(
		applyCmd,
		diffCmd,
		prepareCmd,
		manifestCmd,
		udashCmd,
		showCmd,
		composeCmd,
		versionCmd,
		docsCmd,
		manCmd,
		jsonschemaCmd)
}

func run(command string) error {

	switch command {
	case "apply", "compose/apply":
		udash.Audience = udashOAuthAudience

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
	case "diff", "compose/diff":
		udash.Audience = udashOAuthAudience
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
		err := e.ManifestUpgrade(manifestUpgradeInPlace)
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	case "manifest/pull":
		err := e.PullFromRegistry(manifestPullPolicyReference, disableTLS)
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	case "manifest/push":
		err := e.PushToRegistry(manifestFiles, valuesFiles, secretsFiles, manifestPushPolicyReference, disableTLS, manifestPushPolicyFile, manifestPushFileStore)
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	// Show is deprecated
	case "manifest/show", "show", "compose/show":
		if showClean {
			defer func() {
				if err := e.Clean(); err != nil {
					logrus.Errorf("error in show clean - %s", err)
				}
			}()
		}

		if !showDisablePrepare {
			err := e.Prepare()
			if err != nil {
				logrus.Errorf("%s %s", result.FAILURE, err)
				return err
			}
		}

		err := e.Show()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	case "udash/config":
		configFilePath, err := udash.ConfigFilePath()
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

		logrus.Infof("Config file located at %q", configFilePath)

	case "udash/login":
		err := udash.Login(udashEndpointURL, udashOAuthClientID, udashOAuthIssuer, udashOAuthAudience, udashOAuthAccessToken)
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
			return err
		}

	case "udash/logout":
		err := udash.Logout(udashEndpointURL)
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

func getPolicyFilesFromRegistry() error {

	if slices.Equal(policyReferences, []string{""}) || slices.Equal(policyReferences, []string{}) {
		return nil
	}

	// TODO: To be removed once not experimental anymore
	if !experimental {
		logrus.Warningf("The 'oci registry' feature requires the flag --experimental to be set")
		os.Exit(1)
	}

	for _, policy := range policyReferences {
		policyManifest, policyValues, policySecrets, err := registry.Pull(policy, disableTLS)
		if err != nil {
			return err
		}

		manifestFiles = append(policyManifest, manifestFiles...)
		valuesFiles = append(policyValues, valuesFiles...)
		secretsFiles = append(policySecrets, secretsFiles...)
	}

	return nil
}
