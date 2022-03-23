package text

import (
	"fmt"
	"strings"
)

// MockTextRetriever is a stub implementation of the `TextRetriever` interface to be used in our unit test suites.
// It stores the expected Content and Err
type MockTextRetriever struct {
	Err      error
	Lines    map[string]int
	Contents map[string]string
}

func (mtr *MockTextRetriever) ReadLine(location string, line int) (string, error) {
	contentLines := strings.Split(
		strings.ReplaceAll(mtr.Contents[location], "\r\n", "\n"),
		"\n",
	)
	if len(contentLines) < line {
		return "", fmt.Errorf("I/O error: The file %q only contains %d, less than the specified line %d", location, len(contentLines), line)
	}
	return contentLines[line-1], mtr.Err
}

func (mtr *MockTextRetriever) ReadAll(location string) (string, error) {
	return mtr.Contents[location], mtr.Err
}

func (mtr *MockTextRetriever) WriteLineToFile(lineContent, location string, lineNumber int) error {
	mtr.Lines[location] = lineNumber

	// No "\r\n" to "\n" replacements here as we want to obtain a joined string with the same line delimiter as before
	contentLines := strings.Split(
		mtr.Contents[location],
		"\n",
	)
	// If the original line ends by a "\r", add to to the lineContent (which never ends by one)
	if strings.HasSuffix(contentLines[lineNumber-1], "\r") {
		lineContent = lineContent + "\r"
	}

	contentLines[lineNumber-1] = lineContent

	mtr.Contents[location] = strings.Join(contentLines, "\n")
	return mtr.Err
}

func (mtr *MockTextRetriever) WriteToFile(content string, location string) error {
	mtr.Contents[location] = content
	return mtr.Err
}

func (mtr *MockTextRetriever) FileExists(location string) bool {
	_, exists := mtr.Contents[location]
	return exists
}
