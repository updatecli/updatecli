package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/config"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/yaml"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	conf    config.Config

	rootCmd = &cobra.Command{
		Use:   "updateCli",
		Short: "updateCli is a tool to update yaml key values",
		Long: `
updateCli is a tool to update yaml
key value based on source rule
then validated by conditions`,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "updateCli.yaml", "config file (default is ./updateCli.yaml)")
}

func run() {

	conf.ReadFile(cfgFile)
	conf.Check()

	fmt.Printf("👀\tParsing source:\n")

	switch conf.Source.Kind {
	case "githubRelease":
		fmt.Printf("\tgithubRelease specified\n")
		var spec github.Github
		err := mapstructure.Decode(conf.Source.Spec, &spec)

		if err != nil {
			panic(err)
		}

		conf.Source.Output = spec.GetVersion()

	default:
		fmt.Printf("⚠\tDon't support source kind: %v\n", conf.Source.Kind)
	}

	fmt.Printf("👀\tChecking conditions\n")

	for _, condition := range conf.Conditions {
		switch condition.Kind {
		case "dockerImage":
			var d docker.Docker

			err := mapstructure.Decode(condition.Spec, &d)

			if err != nil {
				panic(err)
			}

			d.Tag = conf.Source.Output

			if ok := d.IsTagPublished(); !ok {
				os.Exit(1)
			}

		default:
			fmt.Printf("\t\t⚠\tDon't support condition: %v\n", condition.Kind)
		}

	}

	fmt.Printf("✍\tUpdating Targets\n")
	for _, target := range conf.Targets {
		switch target.Kind {
		case "yaml":
			var spec yaml.Yaml

			err := mapstructure.Decode(target.Spec, &spec)

			if err != nil {
				fmt.Println(err)
			}

			spec.Update(conf.Source.Output)
		}
	}
}
