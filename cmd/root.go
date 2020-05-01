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

	rootCmd = &cobra.Command{
		Use:   "updateCli",
		Short: "updateCli is a tool to update yaml key values",
		Long: `
updateCli is a tool to update yaml
key value based on source rule
then validated by conditions`,
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
	rootCmd.AddCommand(applyCmd, showCmd, versionCmd)
}

func run(cfg string, command string) {
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
			run(filepath.Join(cfg, file), command)
		}
	} else {
		switch command {
		case "apply":
			err := engine.Run(cfg, valuesFile)
			if err != nil {
				fmt.Printf("\n\u26A0 %s \n\n", err)
			}
		case "show":
			err := engine.Show(cfg, valuesFile)
			if err != nil {
				fmt.Printf("\n\u26A0 %s \n\n", err)
			}
		default:
			fmt.Println("Wrong command")
		}
	}

	fmt.Printf("\n")
}
