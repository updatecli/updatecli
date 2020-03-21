package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/olblak/updateCli/pkg/engine"

	"github.com/spf13/cobra"
)

var (
	cfgFile string

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
		err := engine.Run(cfg)
		if err != nil {
			fmt.Printf("\n\u26A0 %s \n\n", err)
		}
	}

	fmt.Printf("\n\n")
}
