package reports

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Actions []Action

func (ba *Actions) Merge(a *Actions) {
	b := *ba
	act := *a

	for i := range act {
		pipelineFound := false
		for j := range b {
			if act[i].ID == b[j].ID {
				pipelineFound = true
				b[j].Merge(&act[i])
				break
			}
		}
		if !pipelineFound {
			b = append(b, act[i])
		}
	}
}

// String show an action report formatted as a string
func (a *Actions) String() string {
	output := ""

	act := *a

	for i := range act {
		if err := act[i].Sort(); err != nil {
			return ""
		}
		o, err := xml.MarshalIndent(act[i], "", "    ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		if output == "" {
			output = string(o)
			continue
		}
		output = strings.Join([]string{output, string(o)}, "\n")

	}

	return string(output[:])
}

func MergeFromString(old, new string) string {
	var oldReport Actions
	var newReport Actions

	err := unmarshal([]byte(old), &oldReport)
	if err != nil {
		logrus.Errorf("failed parsing old report: %s", err)
		// Return the new report by default if something went wrong
		return new
	}

	err = unmarshal([]byte(new), &newReport)
	if err != nil {
		logrus.Errorf("failed parsing new report: %s", err)
		// Return the new report by default
		return new
	}

	newReport.Merge(&oldReport)
	return newReport.String()
}

// unmarshal parses the htmlReport string and return a struct
func unmarshal(input []byte, a *Actions) (err error) {
	if err := xml.Unmarshal(input, a); err != nil {
		return err
	}

	b := *a
	for i := range b {
		if err := b[i].Sort(); err != nil {
			return err
		}
	}
	return nil
}
