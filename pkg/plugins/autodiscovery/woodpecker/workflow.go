package woodpecker

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	goyaml "go.yaml.in/yaml/v3"
)

var (
	// DefaultFilePatterns specifies the default file patterns to identify Woodpecker workflow files
	DefaultFilePatterns = []string{
		".woodpecker.yml",
		".woodpecker.yaml",
		".woodpecker/*.yml",
		".woodpecker/*.yaml",
		".woodpecker/**/*.yml",
		".woodpecker/**/*.yaml",
	}
)

// Step represents a modern Woodpecker step (steps array format)
type Step struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// PipelineStep represents a legacy Woodpecker pipeline step (pipeline map format)
type PipelineStep struct {
	Image string `yaml:"image"`
}

// Service represents a Woodpecker service
type Service struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// Workflow represents a Woodpecker workflow configuration
type Workflow struct {
	// Modern format: steps as array
	Steps []Step `yaml:"steps"`
	// Legacy format: pipeline as map
	Pipeline map[string]PipelineStep `yaml:"pipeline"`
	// Services
	Services []Service `yaml:"services"`
}

// imageInfo holds information about a discovered image
type imageInfo struct {
	Name     string
	Image    string
	Key      string
	IsLegacy bool
}

// searchWorkflowFiles will look, recursively, for Woodpecker workflow files from a root directory.
func searchWorkflowFiles(rootDir string, filePatterns []string) ([]string, error) {
	workflowFiles := []string{}

	logrus.Debugf("Looking for Woodpecker workflow file(s) in %q", rootDir)

	err := filepath.WalkDir(rootDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", filePath, err)
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, filePath)
		if err != nil {
			logrus.Errorln(err)
			return nil
		}

		for _, pattern := range filePatterns {
			// Handle patterns with directory separators
			match, err := filepath.Match(pattern, relPath)
			if err != nil {
				logrus.Errorln(err)
				continue
			}
			if match {
				workflowFiles = append(workflowFiles, filePath)
				break
			}

			// Also try matching just the filename for simple patterns
			match, err = filepath.Match(pattern, d.Name())
			if err != nil {
				logrus.Errorln(err)
				continue
			}
			if match {
				workflowFiles = append(workflowFiles, filePath)
				break
			}

			// Handle glob patterns with ** for recursive matching
			if strings.Contains(pattern, "**") {
				// Convert ** pattern to check if path matches
				parts := strings.Split(pattern, "**")
				if len(parts) == 2 {
					prefix := parts[0]
					suffix := parts[1]
					if strings.HasPrefix(relPath, prefix) && strings.HasSuffix(relPath, strings.TrimPrefix(suffix, "/")) {
						workflowFiles = append(workflowFiles, filePath)
						break
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Remove duplicates
	seen := make(map[string]bool)
	result := []string{}
	for _, f := range workflowFiles {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}

	logrus.Debugf("%d potential Woodpecker workflow file(s) found", len(result))

	return result, nil
}

// loadWorkflow reads and parses a Woodpecker workflow file
func loadWorkflow(filename string) (*Workflow, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var workflow Workflow
	err = goyaml.Unmarshal(content, &workflow)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

// getImagesFromWorkflow extracts all Docker images from a workflow
func getImagesFromWorkflow(workflow *Workflow) []imageInfo {
	var images []imageInfo

	// Process modern steps format
	for i, step := range workflow.Steps {
		if step.Image != "" {
			name := step.Name
			if name == "" {
				name = fmt.Sprintf("step-%d", i)
			}
			images = append(images, imageInfo{
				Name:     name,
				Image:    step.Image,
				Key:      fmt.Sprintf("$.steps[%d].image", i),
				IsLegacy: false,
			})
		}
	}

	// Process legacy pipeline format
	if len(workflow.Pipeline) > 0 {
		// Sort pipeline keys for deterministic output
		pipelineKeys := make([]string, 0, len(workflow.Pipeline))
		for k := range workflow.Pipeline {
			pipelineKeys = append(pipelineKeys, k)
		}
		sort.Strings(pipelineKeys)

		for _, stepName := range pipelineKeys {
			step := workflow.Pipeline[stepName]
			if step.Image != "" {
				images = append(images, imageInfo{
					Name:     stepName,
					Image:    step.Image,
					Key:      fmt.Sprintf("$.pipeline.%s.image", stepName),
					IsLegacy: true,
				})
			}
		}
	}

	// Process services
	for i, service := range workflow.Services {
		if service.Image != "" {
			name := service.Name
			if name == "" {
				name = fmt.Sprintf("service-%d", i)
			}
			images = append(images, imageInfo{
				Name:     fmt.Sprintf("service-%s", name),
				Image:    service.Image,
				Key:      fmt.Sprintf("$.services[%d].image", i),
				IsLegacy: false,
			})
		}
	}

	return images
}

// discoverWorkflowImageManifests generates Updatecli manifests for Woodpecker workflow files
func (w Woodpecker) discoverWorkflowImageManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := w.rootDir
	// If the spec.RootDir is an absolute path, then it has already been set
	// correctly in the New function.
	if w.spec.RootDir != "" && !path.IsAbs(w.spec.RootDir) {
		searchFromDir = filepath.Join(w.rootDir, w.spec.RootDir)
	}

	foundWorkflowFiles, err := searchWorkflowFiles(searchFromDir, w.filematch)
	if err != nil {
		return nil, err
	}

	for _, foundWorkflowFile := range foundWorkflowFiles {
		relativeWorkflowFile, err := filepath.Rel(w.rootDir, foundWorkflowFile)
		logrus.Debugf("parsing file %q", foundWorkflowFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		dirname := filepath.Dir(relativeWorkflowFile)
		basename := filepath.Base(relativeWorkflowFile)

		workflow, err := loadWorkflow(foundWorkflowFile)
		if err != nil {
			logrus.Debugf("loading Woodpecker workflow from %q: %s", foundWorkflowFile, err)
			continue
		}

		images := getImagesFromWorkflow(workflow)
		if len(images) == 0 {
			continue
		}

		for _, img := range images {
			if img.Image == "" {
				continue
			}

			imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(img.Image)
			if err != nil {
				return nil, fmt.Errorf("parsing image %q: %s", img.Image, err)
			}

			/*
				For the time being, it's not possible to retrieve a list of tags for a specific digest
				without a significant number of API calls. More information on the following issue
				https://github.com/google/go-containerregistry/issues/1297
				until a better solution, we don't handle docker image digests without tags.
			*/
			if imageDigest != "" && imageTag == "" {
				logrus.Debugf("docker digest without specified tag is not supported at the moment for %q", img.Image)
				continue
			}

			// Test if the ignore rule based on path is respected
			if len(w.spec.Ignore) > 0 {
				if w.spec.Ignore.isMatchingRule(
					w.rootDir,
					relativeWorkflowFile,
					img.Image,
				) {
					logrus.Debugf("Ignoring Woodpecker workflow file %q from %q, as matching ignore rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(w.spec.Only) > 0 {
				if !w.spec.Only.isMatchingRule(
					w.rootDir,
					relativeWorkflowFile,
					img.Image,
				) {
					logrus.Debugf("Ignoring Woodpecker workflow file %q from %q, as not matching only rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, imageTag, w.spec.Auths)

			versionFilterKind := w.versionFilter.Kind
			versionFilterPattern := w.versionFilter.Pattern
			versionFilterRegex := w.versionFilter.Regex
			tagFilter := "*"

			registryUsername := ""
			registryPassword := ""
			registryToken := ""

			if sourceSpec != nil {
				versionFilterKind = sourceSpec.VersionFilter.Kind
				versionFilterPattern = sourceSpec.VersionFilter.Pattern
				versionFilterRegex = sourceSpec.VersionFilter.Regex
				tagFilter = sourceSpec.TagFilter

				registryUsername = sourceSpec.Username
				registryPassword = sourceSpec.Password
				registryToken = sourceSpec.Token
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !w.spec.VersionFilter.IsZero() {
				versionFilterKind = w.versionFilter.Kind
				versionFilterPattern, err = w.versionFilter.GreaterThanPattern(imageTag)
				versionFilterRegex = w.versionFilter.Regex
				tagFilter = ""
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					if sourceSpec != nil {
						sourceSpec.VersionFilter.Pattern = "*"
					}
				}
			}

			var tmpl *template.Template
			if w.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigestAndLatest)
				if err != nil {
					return nil, err
				}
			} else if w.digest && sourceSpec == nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigest)
				if err != nil {
					return nil, err
				}
			} else if !w.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateLatest)
				if err != nil {
					return nil, err
				}
			} else {
				logrus.Infoln("No source spec detected")
				return nil, nil
			}

			params := struct {
				ActionID             string
				ImageName            string
				ImageTag             string
				SourceID             string
				RegistryUsername     string
				RegistryPassword     string
				RegistryToken        string
				TargetID             string
				TargetFile           string
				TargetKey            string
				TargetPrefix         string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
				VersionFilterRegex   string
				ScmID                string
			}{
				ActionID:             w.actionID,
				ImageName:            imageName,
				ImageTag:             imageTag,
				SourceID:             img.Name,
				RegistryUsername:     registryUsername,
				RegistryPassword:     registryPassword,
				RegistryToken:        registryToken,
				TargetID:             img.Name,
				TargetFile:           relativeWorkflowFile,
				TargetKey:            img.Key,
				TargetPrefix:         imageName + ":",
				TagFilter:            tagFilter,
				VersionFilterKind:    versionFilterKind,
				VersionFilterPattern: versionFilterPattern,
				VersionFilterRegex:   versionFilterRegex,
				ScmID:                w.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
