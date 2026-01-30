package file

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Spec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type Spec struct {
	//   `file` contains the file path
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	//
	//   remarks:
	//       * `file` is incompatible with `files`
	//       * feel free to look at searchpattern attribute to search for files matching a pattern
	File string `yaml:",omitempty"`
	//   `files` contains the file path(s)
	//
	//   compatible:
	//       * condition
	//       * target
	//
	//   remarks:
	//       * `files` is incompatible with `file`
	//       * feel free to look at searchpattern attribute to search for files matching a pattern
	Files []string `yaml:",omitempty"`
	//   `line` contains the line of the file(s) to manipulate
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	Line int `yaml:",omitempty"`
	//   `content` specifies the content to manipulate
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	Content string `yaml:",omitempty"`
	//   `forcecreate` defines if nonexistent file(s) should be created
	//
	//   compatible:
	//       * target
	ForceCreate bool `yaml:",omitempty"`
	//   `matchpattern` specifies the regexp pattern to match on the file(s)
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	//
	//   remarks:
	//       * For targets: Capture groups (parentheses) in the pattern automatically extract
	//         the current value for changelog generation
	//       * Without capture groups, changelogs show generic "unknown" version changes
	//       * With capture groups, changelogs show actual version changes (e.g., "1.24.5" → "1.25.1")
	//       * Example: `"version":\s*"([\d\.]+)"` captures version numbers for changelogs
	//       * Supports full Go regexp syntax
	MatchPattern string `yaml:",omitempty"`
	//   `replacepattern` specifies the regexp replace pattern to apply on the file(s) content
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	ReplacePattern string `yaml:",omitempty"`
	//   `searchpattern` defines if the MatchPattern should be applied on the file(s) path
	//
	//   If set to true, it modifies the behavior of the `file` and `files` attributes to search for files matching the pattern instead of searching for files with the exact name.
	//   When looking for file path pattern, it requires pattern to match all of name, not just a substring.
	//
	//   The pattern syntax is:
	//
	//   ```
	//       pattern:
	//           { term }
	//       term:
	//           '*'         matches any sequence of non-Separator characters
	//           '?'         matches any single non-Separator character
	//           '[' [ '^' ] { character-range } ']'
	//                       character class (must be non-empty)
	//           c           matches character c (c != '*', '?', '\\', '[')
	//           '\\' c      matches character c
	//
	//       character-range:
	//           c           matches character c (c != '\\', '-', ']')
	//           '\\' c      matches character c
	//           lo '-' hi   matches character c for lo <= c <= hi
	//   ```
	SearchPattern bool `yaml:",omitempty"`
	//   `template` specifies the path to a Go template file to render with source values
	//
	//   compatible:
	//       * target
	//
	//   remarks:
	//       * When using template, the source value is passed as `.source` in the template context
	//       * All Go template functions from sprig are available
	//       * The template file is read and rendered at execution time
	//       * `template` is mutually exclusive with `content`, `line`, `matchpattern`, and `replacepattern`
	//
	//   example:
	//       template: "path/to/template.tmpl"
	Template string `yaml:",omitempty"`
	//
	//	`templateData` specifies additional data to pass to the template
	//
	//	compatible:
	//	    * target
	//
	//	remarks:
	//	    * When using template, the data specified here is passed as additional fields in the template context
	//	    * All Go template functions from sprig are available
	//	    * The template file is read and rendered at execution time
	//	    * `templateData` is optional
	//
	//	example:
	//	    templateData:
	//	        key1: "value1"
	//	        key2: "value2"
	TemplateData map[string]interface{} `yaml:",omitempty"`
}

// File defines a resource of kind "file"
type File struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]fileMetadata // map of file
}

type fileMetadata struct {
	originalPath string
	path         string
	content      string
}

// New returns a reference to a newly initialized File object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec interface{}) (*File, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &File{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
	}

	return newResource, nil
}

func hasDuplicates(values []string) bool {
	uniqueValues := make(map[string]string)
	for _, v := range values {
		uniqueValues[v] = ""
	}

	return len(values) != len(uniqueValues)
}

// initFiles initializes the f.files map
func (f *File) initFiles(workDir string) error {
	f.files = make(map[string]fileMetadata)

	// File as unique element of newResource.files
	if len(f.spec.File) > 0 {
		var foundFiles []string
		var err error
		switch f.spec.SearchPattern {
		case true:
			foundFiles, err = utils.FindFilesMatchingPathPattern(workDir, f.spec.File)
			if err != nil {
				return fmt.Errorf("unable to find file matching %q: %s", f.spec.File, err)
			}
		case false:
			foundFiles = append(foundFiles, f.spec.File)
		}

		for _, filePath := range foundFiles {
			newFile := fileMetadata{
				path:         strings.TrimPrefix(filePath, "file://"),
				originalPath: strings.TrimPrefix(filePath, "file://"),
			}
			f.files[filePath] = newFile
		}
	}

	for _, specFile := range f.spec.Files {
		var foundFiles []string
		var err error

		switch f.spec.SearchPattern {
		case true:
			foundFiles, err = utils.FindFilesMatchingPathPattern(workDir, specFile)
			if err != nil {
				return fmt.Errorf("unable to find files matching %q: %s", f.spec.File, err)
			}

		case false:
			foundFiles = append(foundFiles, f.spec.Files...)
		}

		for _, filePath := range foundFiles {
			newFile := fileMetadata{
				path:         strings.TrimPrefix(filePath, "file://"),
				originalPath: strings.TrimPrefix(filePath, "file://"),
			}
			f.files[filePath] = newFile
		}
	}

	for filePath := range f.files {
		if workDir != "" {
			file := f.files[filePath]
			file.path = joinPathWithWorkingDirectoryPath(file.originalPath, workDir)

			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", file.originalPath, file.path)
			f.files[filePath] = file
		}
	}

	return nil
}

func (f *File) UpdateAbsoluteFilePath(workDir string) {
	for filePath := range f.files {
		if workDir != "" {
			file := f.files[filePath]
			file.path = joinPathWithWorkingDirectoryPath(file.originalPath, workDir)

			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", file.originalPath, file.path)
			f.files[filePath] = file
		}
	}
}

// Validate validates the object and returns an error (with all the failed validation messages) if not valid
func (s *Spec) Validate() error {
	var validationErrors []string

	// Check for all validation
	if len(s.Files) == 0 && len(s.File) == 0 {
		validationErrors = append(validationErrors, "Invalid spec for file resource: both 'file' and 'files' are empty.")
	}
	if len(s.Files) > 0 && len(s.File) > 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.file` and `spec.files` are mutually exclusive")
	}
	if len(s.Files) > 1 && s.Line != 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.files` and `spec.line` are mutually exclusive if there is more than one file")
	}
	if len(s.Files) > 1 && hasDuplicates(s.Files) {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.files` contains duplicated values")
	}
	if s.Line < 0 {
		validationErrors = append(validationErrors, "Line cannot be negative for a file resource.")
	}
	if s.Line > 0 {
		if s.ForceCreate {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.forcecreate` and `spec.line` are mutually exclusive")
		}

		if len(s.MatchPattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.matchpattern` and `spec.line` are mutually exclusive")
		}

		if len(s.ReplacePattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.replacepattern` and `spec.line` are mutually exclusive")
		}
	}
	if len(s.Content) > 0 && len(s.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.replacepattern` and `spec.content` are mutually exclusive")
	}
	if len(s.Template) > 0 {
		if len(s.Content) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.template` and `spec.content` are mutually exclusive")
		}
		if len(s.MatchPattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.template` and `spec.matchpattern` are mutually exclusive")
		}
		if len(s.ReplacePattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.template` and `spec.replacepattern` are mutually exclusive")
		}
		if s.Line > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.template` and `spec.line` are mutually exclusive")
		}
	}
	if len(s.TemplateData) > 0 && len(s.Template) == 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attribute `spec.templateData` is specified but `spec.template` is missing")
	}

	// Return all the validation errors if any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Read puts the content of the file(s) as value of the f.files map if the file(s) exist(s) or log the non existence of the file
func (f *File) Read() error {
	var err error

	// Retrieve files content
	for filePath := range f.files {
		file := f.files[filePath]
		if f.contentRetriever.FileExists(file.path) {
			// Return the specified line if a positive number is specified by user in its manifest
			// Note that in this case we're with a fileCount of 1 (as other cases wouldn't pass validation)
			if f.spec.Line > 0 {
				file.content, err = f.contentRetriever.ReadLine(file.path, f.spec.Line)
				if err != nil {
					return err
				}
			}

			// Otherwise return the textual content
			if f.spec.Line == 0 {
				file.content, err = f.contentRetriever.ReadAll(file.path)
				if err != nil {
					return err
				}
			}
		} else {
			if f.spec.ForceCreate {
				// f.files[filePath] is already set to "", no need for more except logging
				logrus.Infof("Creating a new file at %q", file.originalPath)
			} else {
				if f.spec.Line > 0 {
					return fmt.Errorf("%s The specified line %d of the file %q does not exist", result.FAILURE, f.spec.Line, file.originalPath)
				}
				return fmt.Errorf("%s The specified file %q does not exist. If you want to create it, you must set the attribute 'spec.forcecreate' to 'true'", result.FAILURE, filePath)
			}
		}
		f.files[filePath] = file
	}
	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (f *File) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a new configuration with only the necessary configuration fields
// without any sensitive information or context specific data.
func (f *File) ReportConfig() interface{} {
	return Spec{
		File:           f.spec.File,
		Files:          f.spec.Files,
		Line:           f.spec.Line,
		Content:        f.spec.Content,
		MatchPattern:   f.spec.MatchPattern,
		ReplacePattern: f.spec.ReplacePattern,
		SearchPattern:  f.spec.SearchPattern,
		Template:       f.spec.Template,
		TemplateData:   f.spec.TemplateData,
	}
}
