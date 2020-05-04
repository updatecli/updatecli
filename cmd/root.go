package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/olblak/updateCli/pkg/engine"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	valuesFile string

	e engine.Engine

	rootCmd = &cobra.Command{
		Use:   "updateCli",
		Short: "updateCli is a tool to automate file updates",
		Long: `
updateCli is a tool to automate file updates based on source rule.`,
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

	for _, file := range files {

		switch command {
		case "apply":
			err := e.Apply(file)
			if err != nil {
				fmt.Printf("\n\u26A0 %s \n\n", err)
			}
		case "show":
			err := e.Show(file)
			if err != nil {
				fmt.Printf("\n\u26A0 %s \n\n", err)
			}
		default:
			fmt.Println("Wrong command")
		}
	}

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
