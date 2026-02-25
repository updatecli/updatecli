package cmd

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/registry"
	"github.com/updatecli/updatecli/pkg/core/udash"
	"github.com/updatecli/updatecli/pkg/plugins/utils/ci"

	"github.com/updatecli/updatecli/pkg/core/engine"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/spf13/cobra"
)

var (
	pipelineIds      []string
	labels           []string
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

	logrus.SetOutput(os.Stdout)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "debug", "", false, "Debug Output")
	rootCmd.PersistentFlags().BoolVarP(&experimental, "experimental", "", false, "Enable Experimental mode")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			detectedCi, err := ci.New()
			if err == nil && detectedCi.IsDebug() {
				logrus.Infof("CI pipeline detected in Debug Mode - hence enabling debug mode")
				logrus.SetLevel(logrus.DebugLevel)
			}
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
		pipelineCmd,
		udashCmd,
		showCmd,
		composeCmd,
		versionCmd,
		docsCmd,
		manCmd,
		jsonschemaCmd)
}

func run(command string) error {

	for _, id := range pipelineIds {
		e.Options.PipelineIDs = append(e.Options.PipelineIDs, strings.Split(id, ",")...)
	}

	for _, label := range labels {
		labelsArray := strings.Split(label, ",")

		initLabels := func() {
			if e.Options.Labels == nil {
				e.Options.Labels = make(map[string]string)
			}
		}

		for i := range labelsArray {
			labelKeyValue := strings.SplitN(labelsArray[i], ":", 2)
			if labelKeyValue[0] == "" {
				logrus.Warnf("Ignoring label with empty key: %q", labelsArray[i])
				continue
			}
			switch len(labelKeyValue) {
			case 2:
				initLabels()
				e.Options.Labels[labelKeyValue[0]] = labelKeyValue[1]
			case 1:
				initLabels()
				e.Options.Labels[labelKeyValue[0]] = ""
			}
		}
	}

	switch command {
	case "apply", "compose/apply", "pipeline/apply":
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
	case "diff", "compose/diff", "pipeline/diff":
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

	case "prepare", "pipeline/prepare":
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

	case "manifest/init":

		err := e.Scaffold(manifestInitPolicyRootDir)
		if err != nil {
			logrus.Errorf("%s %s", result.FAILURE, err)
		}

	case "compose/lint":

		err := e.Lint(composeLintPolicyRootDir)
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
		err := e.PushToRegistry(
			manifestFiles,
			valuesFiles,
			secretsFiles,
			manifestPushPolicyReference,
			disableTLS,
			manifestPushPolicyFile,
			manifestPushFileStore,
			manifestPushOverwrite)

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
		err := udash.Login(udashEndpointURL, udashEndpointAPIURL, udashOAuthClientID, udashOAuthIssuer, udashOAuthAudience, udashOAuthAccessToken)
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
