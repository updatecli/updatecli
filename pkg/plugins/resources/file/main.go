package file

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

/*
"file" defines the specification for manipulating any text file.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * set "searchpattern" to treat the path as a pattern.
	//   * a URL such as "https://" is not supported in a target.
	//
	// example:
	//   * file: README.md
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of file paths to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * in a source, "files" accepts at most one entry.
	//   * duplicated entries are rejected.
	//   * set "searchpattern" to treat each path as a pattern.
	//   * a URL such as "https://" is not supported in a target.
	//
	// example:
	//   * files:
	//     - README.md
	//     - docs/README.md
	//
	Files []string `yaml:",omitempty"`
	// "line" defines the line number of the file to manipulate.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   0, which means the whole file
	//
	// remark:
	//   * the first line of the file is line 1.
	//   * "line" cannot be negative.
	//   * "line" is mutually exclusive with "forcecreate", "matchpattern", "replacepattern" and "template".
	//   * "line" cannot be used when "files" holds more than one entry.
	//
	// example:
	//   * line: 3
	//
	Line int `yaml:",omitempty"`
	// "content" defines the content to compare with, or to write to, the file.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	// remark:
	//   * "content" and "replacepattern" are mutually exclusive.
	//   * "content" and "template" are mutually exclusive.
	//   * in a condition, "content" cannot be used together with a "sourceid".
	//
	Content string `yaml:",omitempty"`
	// "forcecreate" creates the file when it does not exist.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * "forcecreate" and "line" are mutually exclusive.
	//
	ForceCreate bool `yaml:",omitempty"`
	// "matchpattern" defines the regular expression matched against the file content.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * it supports the full Go regular expression syntax.
	//   * "matchpattern" is mutually exclusive with "line" and "template".
	//   * in a source, the output is every matching string, one per line.
	//   * in a target, capture groups (parentheses) extract the current value for the changelog.
	//     Without a capture group, the changelog shows a generic "unknown" version change.
	//     With one, it shows the actual change, such as "1.24.5" to "1.25.1".
	//
	// example:
	//   * matchpattern: '"version":\s*"([\d\.]+)"'
	//
	MatchPattern string `yaml:",omitempty"`
	// "replacepattern" defines the regular expression replacement applied to the content matched by "matchpattern".
	//
	// compatible:
	//   * target
	//
	// default:
	//   the output of the associated source, or "content" when set.
	//
	// remark:
	//   * "replacepattern" is mutually exclusive with "content", "line" and "template".
	//   * it only applies when "matchpattern" is set.
	//
	// example:
	//   * replacepattern: '"version": "1.25.1"'
	//
	ReplacePattern string `yaml:",omitempty"`
	// "searchpattern" treats "file" and "files" as path patterns instead of exact paths.
	//
	// The pattern must match the whole path, not just a substring.
	//
	// The pattern syntax is:
	//
	// ```
	//     pattern:
	//         { term }
	//     term:
	//         '*'         matches any sequence of non-Separator characters
	//         '?'         matches any single non-Separator character
	//         '[' [ '^' ] { character-range } ']'
	//                     character class (must be non-empty)
	//         c           matches character c (c != '*', '?', '\\', '[')
	//         '\\' c      matches character c
	//
	//     character-range:
	//         c           matches character c (c != '\\', '-', ']')
	//         '\\' c      matches character c
	//         lo '-' hi   matches character c for lo <= c <= hi
	// ```
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * combined with "matchpattern", files whose content does not match are ignored instead of failing.
	//
	SearchPattern bool `yaml:",omitempty"`
	// "template" defines the path of a Go template file rendered to produce the file content.
	//
	// compatible:
	//   * target
	//
	// remark:
	//   * the source value is available as ".source" in the template.
	//   * every sprig template function is available.
	//   * the template file is read and rendered at execution time.
	//   * "template" is mutually exclusive with "content", "line", "matchpattern" and "replacepattern".
	//
	// example:
	//   * template: path/to/template.tmpl
	//
	Template string `yaml:",omitempty"`
	// "templatedata" defines additional data passed to the template.
	//
	// compatible:
	//   * target
	//
	// remark:
	//   * each entry is available as a field of the template context.
	//   * "templatedata" requires "template".
	//   * the key "source" is reserved and ignored.
	//
	// example:
	// ```
	//   templatedata:
	//     key1: value1
	//     key2: value2
	// ```
	//
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
func (f *File) initFiles(pathResolver pathresolver.Resolver) error {
	f.files = make(map[string]fileMetadata)

	// File as unique element of newResource.files
	if len(f.spec.File) > 0 {
		var foundFiles []string
		var err error
		switch f.spec.SearchPattern {
		case true:
			foundFiles, err = utils.FindFilesMatchingPathPattern(pathResolver.Dir(), f.spec.File)
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
			foundFiles, err = utils.FindFilesMatchingPathPattern(pathResolver.Dir(), specFile)
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

	return f.UpdateAbsoluteFilePath(pathResolver)
}

// UpdateAbsoluteFilePath resolves every file of the f.files map against the path resolver,
// leaving the path the user wrote available as originalPath for reporting.
func (f *File) UpdateAbsoluteFilePath(pathResolver pathresolver.Resolver) error {
	for filePath := range f.files {
		file := f.files[filePath]

		resolvedPath, err := pathResolver.Resolve(file.originalPath)
		if err != nil {
			return err
		}

		if resolvedPath != file.path {
			logrus.Debugf("Relative path detected: changing from %q to %q", file.originalPath, resolvedPath)
		}

		file.path = resolvedPath
		f.files[filePath] = file
	}

	return nil
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
