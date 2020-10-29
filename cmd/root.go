package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/olblak/updateCli/pkg/engine"
	"github.com/olblak/updateCli/pkg/reports"
	"github.com/olblak/updateCli/pkg/result"
	"github.com/olblak/updateCli/pkg/tmp"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	valuesFile string

	e engine.Engine

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
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("\n\u26A0 %s \n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		applyCmd,
		diffCmd,
		showCmd,
		versionCmd)
}

func run(command string) {

	files := GetFiles(e.Options.File)
	reports := reports.Reports{}
	err := tmp.Create()
	if err != nil {
		fmt.Printf("\n\u26A0 %s\n", err)
		os.Exit(1)
	}

	if applyClean && diffClean {
		defer tmp.Clean()
	}

	for _, file := range files {

		switch command {
		case "apply":
			report, err := e.Run(file)
			if err != nil {
				fmt.Printf("\n%s %s \n\n", result.FAILURE, err)
			}
			reports = append(reports, report)
		case "diff":
			report, err := e.Run(file)
			if err != nil {
				fmt.Printf("\n%s %s \n\n", result.FAILURE, err)
			}
			reports = append(reports, report)
		case "show":
			err := e.Show(file)
			if err != nil {
				fmt.Printf("\n%s %s \n\n", result.FAILURE, err)
			}
		default:
			fmt.Println("Wrong command")
		}
	}

	reports.Show()
	reports.Summary()
	fmt.Printf("\n")
}

// GetFiles return an array with every valid files
func GetFiles(root string) (files []string) {

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("\n\u26A0 File %s: %s\n", path, err)
			os.Exit(1)
		}
		if info.Mode().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	return files
}
