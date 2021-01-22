package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// The docs command is a hidden sub-command that generates the documentation for www.updatecli.io

var (
	docPrepender = func(s string) string {

		header := "---\n"
		date := time.Now()
		date.Format(time.RFC3339)

		name := filepath.Base(s)
		name = strings.ReplaceAll(name, "_", " ")
		name = strings.TrimSuffix(name, ".md")

		header = header + fmt.Sprintf("title: %v\n", name)
		header = header + fmt.Sprintf("description: Documentation for the command `%v`\n", name)
		header = header + fmt.Sprintf("lead: Documentation for the command `%v`\n", name)
		header = header + fmt.Sprintf("date: %v\n", date.Format(time.RFC3339))
		header = header + fmt.Sprintf("lastmod: %v\n", date.Format(time.RFC3339))
		header = header + "draft: false\n"
		header = header + "images: []\n"
		header = header + "menu:\n  docs:\n    parent: \"commands\"\n"
		header = header + "weight: 130\n"
		header = header + "toc: true\n"
		header = header + "---\n\n"

		return header
	}

	linkhandler = func(s string) string {
		s = strings.ToLower(s)
		s = strings.TrimSuffix(s, ".md")
		s = "/docs/commands/" + s

		return s
	}

	docsDir string

	docsCmd = &cobra.Command{
		Use:    "docs",
		Hidden: true,
		Short:  "Generate updatecli documentation",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.DisableAutoGenTag = true
			err := doc.GenMarkdownTreeCustom(
				rootCmd,
				docsDir,
				docPrepender,
				linkhandler)
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
	docsCmd.Flags().StringVarP(&docsDir, "docs", "d", "./docs", "Specify the directory where to generate documentation files (default: './docs')")
}
