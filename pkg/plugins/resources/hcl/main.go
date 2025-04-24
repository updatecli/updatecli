package hcl

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/minamijoyo/hcledit/editor"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

type Hcl struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]file // map of file paths to file contents
}

type file struct {
	originalFilePath string
	filePath         string
	content          string
}

func New(spec interface{}) (*Hcl, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &Hcl{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
	}

	newResource.files = make(map[string]file)
	// File as unique element of newResource.files
	if len(newResource.spec.File) > 0 {
		filePath := strings.TrimPrefix(newResource.spec.File, "file://")
		newResource.files[filePath] = file{
			originalFilePath: filePath,
			filePath:         filePath,
		}
	}
	// Files
	for _, filePath := range newResource.spec.Files {
		filePath := strings.TrimPrefix(filePath, "file://")
		newResource.files[filePath] = file{
			originalFilePath: filePath,
			filePath:         filePath,
		}
	}

	return newResource, nil
}

func (h *Hcl) Query(resourceFile file) (string, error) {
	query := h.spec.Path

	sink := editor.NewAttributeGetSink(query, true)

	inStream := strings.NewReader(resourceFile.content)
	outStream := new(bytes.Buffer)
	err := editor.DeriveStream(inStream, outStream, resourceFile.filePath, sink)
	if err != nil {
		return "", err
	}

	// When a value isn't found hcledit returns an empty byte array,
	// where as result always has \n appended so will be > 0
	if outStream.Len() == 0 {
		err := fmt.Errorf("%s cannot find value for path %q from file %q",
			result.FAILURE,
			query,
			resourceFile.originalFilePath)
		return "", err
	}

	queryOutput := outStream.String()
	queryOutput = strings.TrimSpace(queryOutput)
	queryOutput = strings.Trim(queryOutput, "\"")
	return queryOutput, nil
}

func (h *Hcl) Apply(filePath string, valueToWrite string) error {
	query := h.spec.Path

	if _, err := strconv.Atoi(valueToWrite); err != nil {
		valueToWrite = fmt.Sprintf(`"%s"`, valueToWrite)
	}

	resourceFile := h.files[filePath]

	filter := editor.NewAttributeSetFilter(query, valueToWrite)
	inStream := strings.NewReader(resourceFile.content)
	outStream := new(bytes.Buffer)
	err := editor.EditStream(inStream, outStream, resourceFile.filePath, filter)
	if err != nil {
		return err
	}

	resourceFile.content = outStream.String()

	h.files[filePath] = resourceFile

	return nil
}

// Read puts the content of the file(s) as value of the y.files map if the file(s) exist(s) or log the non existence of the file
func (h *Hcl) Read() error {
	var err error

	// Retrieve files content
	for filePath := range h.files {
		f := h.files[filePath]
		if h.contentRetriever.FileExists(f.filePath) {
			f.content, err = h.contentRetriever.ReadAll(f.filePath)
			if err != nil {
				return err
			}
			h.files[filePath] = f

		} else {
			return fmt.Errorf("%s The specified file %q does not exist", result.FAILURE, f.filePath)
		}
	}
	return nil
}

func (h *Hcl) UpdateAbsoluteFilePath(workDir string) {
	for filePath := range h.files {
		if workDir != "" {
			f := h.files[filePath]
			f.filePath = utils.JoinFilePathWithWorkingDirectoryPath(f.originalFilePath, workDir)
			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", f.originalFilePath, f.filePath)
			h.files[filePath] = f
		}
	}
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (h *Hcl) Changelog(from, to string) *result.Changelogs {
	return nil
}

// CleanConfig return a new configuration without any sensitive information
// or context specific data.
func (h *Hcl) CleanConfig() interface{} {
	return Spec{
		File:  h.spec.File,
		Files: h.spec.Files,
		Path:  h.spec.Path,
		Value: h.spec.Value,
	}
}
