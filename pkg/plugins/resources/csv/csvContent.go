package csv

import (
	"encoding/csv"
	"fmt"
	"strings"

	das "github.com/tomwright/dasel"

	"github.com/tomwright/dasel/storage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

// csvContent is *** of the dasel FileContent
type csvContent struct {
	dasel.FileContent
	csvDocument storage.CSVDocument
	comma       rune
	comment     rune
}

func (c *csvContent) Read(rootDir string) error {

	c.FilePath = dasel.JoinPathWithWorkingDirectoryPath(c.FilePath, rootDir)

	// Test at runtime if a file exist
	if !c.ContentRetriever.FileExists(c.FilePath) {
		return fmt.Errorf("the CSV file %q does not exist", c.FilePath)
	}

	textContent, err := c.ContentRetriever.ReadAll(c.FilePath)
	if err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(textContent))

	r.Comma = c.comma
	r.Comment = c.comment

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

	c.DaselNode = das.New(c.csvDocument.Documents())

	return nil
}
