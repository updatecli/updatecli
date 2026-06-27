package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// The docs command is a hidden sub-command that generates the documentation for www.updatecli.io

var (
	docPrepender = func(s string) string {

		header := "---\n"

		name := filepath.Base(s)
		name = strings.ReplaceAll(name, "_", " ")
		name = strings.TrimSuffix(name, ".md")

		header = header + fmt.Sprintf("title: %v\n", name)
		header = header + fmt.Sprintf("description: Documentation for the command `%v`\n", name)
		header = header + fmt.Sprintf("lead: Documentation for the command `%v`\n", name)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			rootCmd.DisableAutoGenTag = true

			before, err := snapshotDir(docsDir)
			if err != nil {
				return err
			}

			if err = doc.GenMarkdownTreeCustom(rootCmd, docsDir, docPrepender, linkhandler); err != nil {
				return err
			}

			after, err := snapshotDir(docsDir)
			if err != nil {
				return err
			}

			for path, hash := range after {
				if prev, exists := before[path]; !exists || prev != hash {
					logrus.Infof("Documentation files updated in %q", docsDir)
					os.Exit(2)
				}
			}

			logrus.Infof("Documentation files already up to date in %q", docsDir)
			return nil
		},
	}
)

func init() {
	docsCmd.Flags().StringVarP(&docsDir, "docs", "d", "./docs", "Specify the directory where to generate documentation files")
}

// snapshotDir returns a map of file paths (relative to dir) to their SHA-256 hash.
// If the directory does not exist yet, an empty map is returned.
func snapshotDir(dir string) (map[string][sha256.Size]byte, error) {
	hashes := make(map[string][sha256.Size]byte)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return hashes, nil
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		h := sha256.New()
		if _, err = io.Copy(h, f); err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		hashes[rel] = [sha256.Size]byte(h.Sum(nil))
		return nil
	})

	return hashes, err
}
