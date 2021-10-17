package text

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type TextRetriever interface {
	ReadLine(location string, line int) (string, error)
	ReadAll(location string) (string, error)
	WriteToFile(content string, location string) error
	WriteLineToFile(lineContent, location string, lineNumber int) error
	FileExists(location string) bool
}

type Text struct{}

// readFromURL reads a text content from an http/https url
func readFromURL(url string, line int) (string, error) {
	// #nosec G107 // url is always "user-defined" so it's tainted by nature
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Only retrieve the specified line in memory if specified
	if line > 0 {
		return getLine(bufio.NewReader(resp.Body), line), nil
	}

	// Otherwise retrieve the whole file content. Can be heavy.
	bodyContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyContent), nil
}

// readFromFile returns the full content of a file from the specified location
// such as a "file://" URL or a standard filesystem path (default)
func readFromFile(location string, line int) (string, error) {
	// Only retrieve the specified line in memory if specified
	if line > 0 {
		file, err := os.Open(location)
		if err != nil {
			return "", err
		}
		defer file.Close()

		return getLine(bufio.NewReader(file), line), nil
	}

	// Otherwise retrieve the whole file content. Can be heavy.
	fileContent, err := ioutil.ReadFile(location)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}

func getLine(reader io.Reader, line int) string {
	// Iterate over the lines in a buffered text content
	// until you find the correct (to avoid loading the file in memory in one shot)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	// Line number are 1-indexed
	currentLine := 1
	for scanner.Scan() {
		if currentLine == line {
			return scanner.Text()
		}
		currentLine++
	}

	return ""
}

// ReadAll reads text content from a location (URL or filepath).
// The location accepts multiple input strings: starting with either "http://",
// "https://", or file url "file://" or filepath (default)
func (t *Text) ReadAll(location string) (string, error) {
	if IsURL(location) {
		content, err := readFromURL(location, 0)
		if err != nil {
			return "", err
		}

		return content, err
	}

	// If it's not an URL, then it's a file path!
	filepath := location
	if strings.HasPrefix(location, "file://") {
		filepath = strings.TrimPrefix(filepath, "file://")
	}

	content, err := readFromFile(filepath, 0)
	if err != nil {
		return "", err
	}

	return content, err
}

// ReadLine reads the specified line of text from the specified location (URL or filepath).
// The location accepts multiple input strings: starting with either "http://",
// "https://", or file url "file://" or filepath (default)
func (t *Text) ReadLine(location string, line int) (string, error) {
	if IsURL(location) {
		content, err := readFromURL(location, line)
		if err != nil {
			return "", err
		}

		return content, err
	}

	// If it's not an URL, then it's a file path!
	filepath := location
	if strings.HasPrefix(location, "file://") {
		filepath = strings.TrimPrefix(filepath, "file://")
	}

	content, err := readFromFile(filepath, line)
	if err != nil {
		return "", err
	}

	return content, err
}

// Diff return a diff like string, comparing string A and string B
func Diff(a, b string) (result string) {

	for _, line := range strings.Split(a, "\n") {
		result = result + "< " + line + "\n"
	}

	result = result + "---\n"

	for _, line := range strings.Split(b, "\n") {
		result = result + "> " + line + "\n"
	}
	return result

}

// Show return a string where each line start with a tabulation
// to increase visibility
func Show(content string) (result string) {

	result = result + "---\n"

	for _, line := range strings.Split(content, "\n") {
		result = result + "| " + line + "\n"
	}

	result = result + "---\n"

	return result

}

// IsURL tests if a string is a valid http URL
func IsURL(location string) bool {
	_, err := os.Stat(location)
	if err == nil {
		// If "location" exists is not an exitring file, then let's try an URL
		// Note that we do not check error type: URL parsing will cover edge cases
		return false
	}

	url, err := url.ParseRequestURI(location)
	if err != nil {
		return false
	}

	address := net.ParseIP(url.Host)

	if address == nil {
		return strings.Contains(url.Host, ".")
	}

	return true
}

// WriteToFile write a string to a file
func (t *Text) WriteToFile(content string, location string) error {

	// If path is not absolute then we specify it to the current directory
	if !filepath.IsAbs(location) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		location = filepath.Join(wd, location)
	}

	file, err := os.Create(location)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil

}

// WriteLineToFile writes 'lineContent' to the line number 'lineNumber' in the file at 'location'
func (t *Text) WriteLineToFile(lineContent, location string, lineNumber int) (err error) {

	// Get actual content of the file
	fileContent, err := t.ReadAll(location)
	if err != nil {
		return err
	}

	// Open the file
	file, err := os.OpenFile(location, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the "write" data buffer
	datawriter := bufio.NewWriter(file)

	// Iterate over the lines in a buffered text content
	// until you find the correct line to write to
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	scanner.Split(bufio.ScanLines)

	// Line number are 1-indexed
	currentLine := 1
	for scanner.Scan() {
		if currentLine == lineNumber {
			_, _ = datawriter.WriteString(lineContent + "\n")
		} else {
			_, _ = datawriter.WriteString(scanner.Text() + "\n")
		}
		currentLine++
	}

	datawriter.Flush()
	return nil
}

func (t *Text) FileExists(location string) bool {
	// Check that the specified exists or exit with error
	if _, err := os.Stat(location); err != nil {
		return false
	}

	return true
}
