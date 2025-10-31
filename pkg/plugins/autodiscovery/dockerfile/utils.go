package dockerfile

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser/keywords"
)

const (
	// extractVariableRegex defines the regular expression used to extract Dockerfile ARGS from value
	//extractVariableRegex string = `(.*)\$\{(.*)\}(.*)`

	FromInstruction = "FROM"
	ArgInstruction  = "ARG"
)

// searchDockerfiles will look, recursively, for every files matching the default pattern "Dockerfile" from a root directory.
func searchDockerfiles(rootDir string, files []string) ([]string, error) {

	dockerfiles := []string{}

	logrus.Debugf("Looking for Dockerfile(s) in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			match, err := filepath.Match(f, info.Name())
			if err != nil {
				logrus.Errorln(err)
				continue
			}
			if match {
				dockerfiles = append(dockerfiles, path)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d potential Dockerfile(s) found", len(dockerfiles))

	return dockerfiles, nil
}

// parseDockerfile reads a Dockerfile for information that could be automated such as ARG or FROM
func parseDockerfile(filename string) ([]keywords.FromToken, map[string]keywords.SimpleTokens, error) {

	if _, err := os.Stat(filename); err != nil {
		return nil, nil, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer v.Close()

	dockerfileContent, err := io.ReadAll(v)
	if err != nil {
		return nil, nil, err
	}

	instructions, args, err := searchInstructions(dockerfileContent)
	if err != nil {
		return nil, nil, err
	}

	if len(instructions) == 0 {
		return nil, nil, nil
	}

	return instructions, args, nil
}

// Search for both a Dockerfile instruction required to define the update manifest.
// While the dockerfile instruction is not case sensitive, its value is
func searchInstructions(dockerfileContent []byte) ([]keywords.FromToken, map[string]keywords.SimpleTokens, error) {
	argParser, err := simpletextparser.NewSimpleTextDockerfileParser(map[string]string{
		"keyword": "ARG",
		"matcher": "",
	})
	if err != nil {
		return nil, nil, err
	}

	fromParser, err := simpletextparser.NewSimpleTextDockerfileParser(map[string]string{
		"keyword": "FROM",
		"matcher": "",
	})
	if err != nil {
		return nil, nil, err
	}
	argInstructions := argParser.GetInstructionTokens(dockerfileContent)
	// let' s construct a map of those
	args := make(map[string]keywords.SimpleTokens)
	for _, rawArg := range argInstructions {
		arg, ok := rawArg.(keywords.SimpleTokens)
		if !ok {
			continue
		}
		args[arg.Name] = arg
	}
	fromInstructions := fromParser.GetImages(dockerfileContent)
	return fromInstructions, args, nil
}
