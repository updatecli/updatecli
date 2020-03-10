package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/config"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/maven"
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
			run(cfgFile)
		},
	}
)

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("\n\u26A0 %s \n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./updateCli.yaml", "config file (default is ./updateCli.yaml)")
}

func run(cfg string) {
	fileInfo, err := os.Stat(cfg)
	if err != nil {
		fmt.Printf("\n\u26A0 %s \n", err)
	}

	if os.IsNotExist(err) {
		fmt.Println(err)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		fmt.Println("Directory configuration provided")
		dir, err := os.Open(cfg)
		defer dir.Close()
		if err != nil {
			fmt.Printf("\n\u26A0 %s \n", err)
		}
		files, err := dir.Readdirnames(-1)
		fmt.Printf("Detected configuration Files: %v \n", files)
		for _, file := range files {
			run(filepath.Join(cfg, file))
		}
	} else {
		err := engine(cfg)
		if err != nil {
			fmt.Printf("\n\u26A0 %s \n", err)
		}
	}
}

func engine(cfgFile string) error {

	_, basename := filepath.Split(cfgFile)
	cfgFileName := strings.TrimSuffix(basename, filepath.Ext(basename))

	fmt.Printf("\n\n%s\n", strings.Repeat("#", len(cfgFileName)+4))
	fmt.Printf("# %s #\n", strings.ToTitle(cfgFileName))
	fmt.Printf("%s\n\n", strings.Repeat("#", len(cfgFileName)+4))

	conf.ReadFile(cfgFile)

	conf.Check()

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Source"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Source")+1))

	switch conf.Source.Kind {
	case "githubRelease":
		var spec github.Github
		err := mapstructure.Decode(conf.Source.Spec, &spec)

		if err != nil {
			panic(err)
		}

		conf.Source.Output = spec.GetVersion()

	case "maven":
		var spec maven.Maven
		err := mapstructure.Decode(conf.Source.Spec, &spec)

		if err != nil {
			panic(err)
		}

		conf.Source.Output = spec.GetVersion()

	case "dockerDigest":
		fmt.Printf("dockerDigest specified\n")
		var spec docker.Docker
		err := mapstructure.Decode(conf.Source.Spec, &spec)

		fmt.Printf("Looking for %s:%s digest\n\n", spec.Image, spec.Tag)

		if err != nil {
			panic(err)
		}

		conf.Source.Output = spec.GetVersion()

	default:
		fmt.Printf("⚠ Don't support source kind: %v\n", conf.Source.Kind)
	}

	fmt.Printf("\n\n%s:\n", strings.ToTitle("conditions"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("conditions")+1))

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
				return fmt.Errorf("skipping: condition not met")
			}
			fmt.Printf("\n")
		case "maven":
			var m maven.Maven

			err := mapstructure.Decode(condition.Spec, &m)

			if err != nil {
				panic(err)
			}

			m.Version = conf.Source.Output

			if ok := m.IsTagPublished(); !ok {
				return fmt.Errorf("skipping: condition not met")
			}
			fmt.Printf("\n")

		default:
			fmt.Printf("⚠ Don't support condition: %v\n", condition.Kind)
		}

	}

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Targets"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Targets")+1))

	for _, target := range conf.Targets {
		switch target.Kind {
		case "yaml":
			var spec yaml.Yaml

			err := mapstructure.Decode(target.Spec, &spec)

			if err != nil {
				fmt.Println(err)
			}

			spec.Update(conf.Source.Output)

		default:
			fmt.Printf("⚠ Don't support target: %v\n", target.Kind)
		}

		fmt.Printf("\n\n")
	}
	return nil
}
