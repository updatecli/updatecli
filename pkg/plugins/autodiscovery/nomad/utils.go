package nomad

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	terraformUtils "github.com/updatecli/updatecli/pkg/plugins/resources/terraform"

	"github.com/sirupsen/logrus"
)

// searchNomadFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchNomadFiles(rootDir string, filePatterns []string) ([]string, error) {

	//results := []nomadDockerSpec{}

	potentialNomadFiles := []string{}

	logrus.Debugf("Looking for Nomad file(s) in %q", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if !d.IsDir() {
			for _, f := range filePatterns {
				match, err := filepath.Match(f, d.Name())
				if err != nil {
					logrus.Errorln(err)
					continue
				}
				if match {
					potentialNomadFiles = append(potentialNomadFiles, path)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return potentialNomadFiles, nil
}

// getNomadDockerSpecFromFile reads a Nomad files for information that could be automated
func getNomadDockerSpecFromFile(filename string) ([]nomadDockerSpec, error) {

	results := []nomadDockerSpec{}

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return nil, err
	}

	hclfile, err := terraformUtils.ParseHcl(string(content), filename)
	if err != nil {
		return nil, err
	}

	variables := map[string]string{}

	for _, block := range hclfile.Body().Blocks() {
		if block.Type() == "variable" {
			variableName := block.Labels()[0]
			variableTypeAttribute := block.Body().GetAttribute("type")
			variableDefaultAttribute := block.Body().GetAttribute("default")
			if variableDefaultAttribute != nil && variableTypeAttribute != nil {
				variableDefaultValue := strings.Trim(string(variableDefaultAttribute.Expr().BuildTokens(nil).Bytes()), " \"")
				variableTypeValue := strings.Trim(string(variableTypeAttribute.Expr().BuildTokens(nil).Bytes()), " \"")

				if variableTypeValue == "string" {
					variables[variableName] = variableDefaultValue
				}
			}
		}
		if block.Type() == "job" {
			jobName := block.Labels()[0]
			for _, groupBlock := range block.Body().Blocks() {
				if groupBlock.Type() == "group" {
					groupName := groupBlock.Labels()[0]
					for _, taskBlock := range groupBlock.Body().Blocks() {
						if taskBlock.Type() == "task" {
							taskName := taskBlock.Labels()[0]
							driverAttribute := taskBlock.Body().GetAttribute("driver")

							// If nil then it means that the driver is not set
							// so we skip this task and continue to the next one
							if driverAttribute == nil {
								continue
							}

							driverValue := strings.Trim(string(driverAttribute.Expr().BuildTokens(nil).Bytes()), " \"")

							if driverValue == "docker" || driverValue == "podman" || driverValue == "containerd-driver" {
								for _, configBlock := range taskBlock.Body().Blocks() {
									if configBlock.Type() == "config" {
										imageAttribute := configBlock.Body().GetAttribute("image")
										if imageAttribute != nil {
											imageValue := string(imageAttribute.Expr().BuildTokens(nil).Bytes())
											imageValue = strings.Trim(imageValue, " \"")

											variableName, err := getVariableName(imageValue)
											if err != nil {
												logrus.Warningf("%s", err)
												continue
											}

											path := fmt.Sprintf("job.%s.group.%s.task.%s.config.image", jobName, groupName, taskName)
											value := imageValue
											if variableName != "" {
												value, err = interpolateVariableName(imageValue, variables[variableName])
												if err != nil {
													logrus.Warningf("%s", err)
													continue
												}
												path = fmt.Sprintf("variable.%s.default", variableName)
											}

											results = append(results, nomadDockerSpec{
												File:      filename,
												Value:     value,
												JobName:   jobName,
												GroupName: groupName,
												TaskName:  taskName,
												Path:      path,
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return results, nil
}

// getVariableName extracts the variable name from a string
func getVariableName(input string) (string, error) {

	if strings.Count(input, "${") > 1 {
		return "", fmt.Errorf("multiple variable detected in image value %q", input)
	}

	re := regexp.MustCompile(`\${\s*var\.(\w+)\s*}`)

	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		variableName := matches[1]
		return variableName, nil
	}
	return "", nil
}

// interpolateVariableName replaces the variable name in a string with the provided variable name
func interpolateVariableName(input string, variableName string) (string, error) {
	re := regexp.MustCompile(`\${\s*var\.(\w+)\s*}`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return strings.ReplaceAll(input, matches[0], variableName), nil
	}
	return "", fmt.Errorf("no variable name found in %q", input)
}
