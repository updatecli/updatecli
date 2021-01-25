package file

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ReadFromURL read a file form a http/https url then return its data as an array of byte.
func ReadFromURL(file string) (data []byte, err error) {
	if IsURL(file) {
		// #nosec G107
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// ReadFromFile read a file then return its data as an array of byte.
func ReadFromFile(file string) (data []byte, err error) {
	// If path is not absolute then we specify it to the current directory
	if !filepath.IsAbs(file) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		file = filepath.Join(wd, file)
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	data, err = ioutil.ReadAll(f)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// HasPrefix test if a filename uses a prefix
func HasPrefix(filename string, prefixes []string) bool {

	for _, prefix := range prefixes {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}

	return false

}

// Read read file from a location then return
// an array of byte. The location accepts multiple input
// http/https urls "https://", file url "file://", or a simple file
func Read(filename, workingDir string) (data []byte, err error) {

	if strings.HasPrefix(filename, "https://") ||
		strings.HasPrefix(filename, "http://") {
		data, err = ReadFromURL(filename)

		if err != nil {
			return nil, err
		}
		return data, err

	} else if strings.HasPrefix(filename, "file://") {
		filename = strings.TrimPrefix(filename, "file://")
	}

	data, err = ReadFromFile(filepath.Join(workingDir, filename))

	if err != nil {
		return nil, err
	}

	return data, err
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

// WriteToFile write a string to a file
func WriteToFile(content string, filename string) (err error) {

	// If path is not absolute then we specify it to the current directory
	if !filepath.IsAbs(filename) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filename = filepath.Join(wd, filename)
	}

	file, err := os.Create(filename)
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

// IsURL tests if a string is a valid http URL
func IsURL(str string) bool {
	url, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	address := net.ParseIP(url.Host)

	if address == nil {
		return strings.Contains(url.Host, ".")
	}

	return true
}
