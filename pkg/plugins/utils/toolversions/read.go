package toolversions

import (
	"bufio"
	"fmt"
	"strings"
)

// Read reads the content of a file after runtime validation
func (f *FileContent) Read(rootDir string) error {

	f.FilePath = JoinPathWithWorkingDirectoryPath(f.FilePath, rootDir)

	if !f.ContentRetriever.FileExists(f.FilePath) {
		return fmt.Errorf("file %q does not exist", f.FilePath)
	}

	textContent, err := f.ContentRetriever.ReadAll(f.FilePath)

	if err != nil {
		return err
	}

	entries, err := readToolVersions(textContent)
	if err != nil {
		return err
	}

	if entries != nil {
		f.Entries = entries
		return nil
	}

	return ErrToolVersionsFailedParsingByteFormat

}

// readToolVersions given the content of the .tool-versions file and returns a list of entries.
func readToolVersions(content string) ([]Entry, error) {

	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // Trim leading and trailing white spaces
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line) // Use Fields to automatically handle multiple spaces and irregular spacing
		if len(parts) == 2 {
			entries = append(entries, Entry{Key: parts[0], Value: parts[1]})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// If no entries were added, return an explicitly initialized empty slice instead of nil.
	if len(entries) == 0 {
		return []Entry{}, nil
	}

	return entries, nil
}
