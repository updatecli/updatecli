package reports

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/tmp"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/time"
)

// ExportToYAML exports the report to a YAML file in the temporary report directory.
// The filename is based on the report ID and the current timestamp.
func (r *Report) ExportToYAML(reportDir string, hideTimestamp bool) (string, error) {
	var err error

	if r == nil {
		return "", errors.New("report doesn't exist")
	}

	if reportDir == "" {
		reportDir, err = tmp.InitReport()
		if err != nil {
			return "", fmt.Errorf("init report directory: %w", err)
		}
	}

	reportDir = filepath.Join(reportDir, r.ID)

	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		err := os.MkdirAll(reportDir, 0755)
		if err != nil {
			return "", err
		}
	}

	reportFilename := filepath.Join(
		reportDir,
		fmt.Sprintf("%s.yaml", time.Now().Format("20060101150405")),
	)

	byteReport, err := yaml.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("marshal report to YAML: %w", err)
	}

	if err := os.WriteFile(reportFilename, byteReport, fs.ModePerm); err != nil {
		return "", err
	}

	return reportFilename, nil
}
