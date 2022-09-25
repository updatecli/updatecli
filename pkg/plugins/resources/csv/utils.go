package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomwright/dasel/storage"
)

// joinPathwithworkingDirectoryPath To merge File path with current working dire, unless file is an http url
func joinPathWithWorkingDirectoryPath(fileName, workingDir string) string {
	if workingDir == "" ||
		filepath.IsAbs(fileName) ||
		strings.HasPrefix(fileName, "https://") ||
		strings.HasPrefix(fileName, "http://") {
		return fileName
	}

	return filepath.Join(workingDir, fileName)
}

func (c *CSV) ReadFromFile() error {

	// Test at runtime if a file exist
	if !c.contentRetriever.FileExists(c.spec.File) {
		return fmt.Errorf("the CSV file %q does not exist", c.spec.File)
	}

	if err := c.Read(); err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(c.currentContent))

	r.Comma = c.spec.Comma
	r.Comment = c.spec.Comment

	res := make([]map[string]interface{}, 0)
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read csv file: %w", err)
	}

	if len(records) == 0 {
		return nil
	}
	var headers []string
	for i, row := range records {
		if i == 0 {
			headers = row
			continue
		}
		rowRes := make(map[string]interface{})
		allEmpty := true
		for index, val := range row {
			if val != "" {
				allEmpty = false
			}
			rowRes[headers[index]] = val
		}
		if !allEmpty {
			res = append(res, rowRes)
		}
	}
	c.csvDocument = storage.CSVDocument{
		Value:   res,
		Headers: headers,
	}

	return nil
}

func (c *CSV) WriteToFile(resourceFile string) error {
	newFile, err := os.Create(resourceFile)
	if err != nil {
		return fmt.Errorf("could not write to file : %w", err)
	}

	defer newFile.Close()

	writer := csv.NewWriter(newFile)

	writer.Comma = c.spec.Comma

	// Iterate through the rows and write the output.
	for i, r := range c.csvDocument.Value {
		if i == 0 {
			if err := writer.Write(c.csvDocument.Headers); err != nil {
				return fmt.Errorf("could not write headers: %w", err)
			}
		}

		values := make([]string, 0)
		for _, header := range c.csvDocument.Headers {
			val, ok := r[header]
			if !ok {
				val = ""
			}
			values = append(values, fmt.Sprint(val))
		}

		if err := writer.Write(values); err != nil {
			return fmt.Errorf("could not write headers: %w", err)
		}

		writer.Flush()
	}

	return nil
}
