package dockerfile

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

const (
	// extractVariableRegex defines the regular expression used to extract Dockerfile ARGS from value
	extractVariableRegex string = `(.*)\$\{(.*)\}(.*)`

	FromInstruction = "FROM"
	ArgInstruction  = "ARG"
)

var (
	// The compiled version of the regex created at init() is cached here so it
	// only needs to be created once.
	regexVariableName *regexp.Regexp
)

func init() {
	regexVariableName = regexp.MustCompile(extractVariableRegex)
}

// instruction defines the struct holding instruction information used to craft the dockerfile manifest
type instruction struct {
	// name define the instruction name such as FROM or ARG
	name string
	// value define the main instruction value based on the type.
	value string
	// arch stores the arch information for a specific FROM. Fetch from the arch flag
	arch string
	// image stores the full image including tag needed to update either the FROM or ARG instruction
	image string
	// trimArgPrefix is only used when we need to update an ARG value based on a FROM information.
	trimArgPrefix string
	// trimArgSuffix is only used when we need to update an ARG value based on a FROM information.
	trimArgSuffix string
}

// searchDockerfiles will look, recursively, for every files matching the default pattern "Dockerfile" from a root directory.
func searchDockerfiles(rootDir string, files []string) ([]string, error) {

	dockerfiles := []string{}

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
func parseDockerfile(filename string) ([]instruction, error) {

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer v.Close()

	dockerfileContent, err := io.ReadAll(v)
	if err != nil {
		return nil, err
	}

	instructions, err := searchInstructions(dockerfileContent)
	if err != nil {
		return nil, err
	}

	if len(instructions) == 0 {
		return nil, nil
	}

	return instructions, nil
}

// Search for both a Dockerfile instruction required to define the update manifest.
// While the dockerfile instruction is not case sensitive, its value is
func searchInstructions(dockerfileContent []byte) ([]instruction, error) {
	instructions := []instruction{}

	data, err := parser.Parse(bytes.NewReader(dockerfileContent))
	if err != nil {
		return nil, err
	}

	args := map[string]string{}

	i := 0
	node := data.AST
	for _, n := range node.Children {
		switch n.Value {
		case FromInstruction:
			value := searchFromValue(n)
			i++

			// Parse Platform flag to extract a potential arch
			platform := searchInstructionFlag("platform", n.Flags)
			arch := ""
			switch regexVariableName.Match([]byte(platform)) {
			case true:
				_, argName, _ := extractArgName(platform)

				if _, found := args[argName]; !found {
					logrus.Debugf("No arg key %q found", argName)
				} else {
					arch = parsePlatform(strings.ReplaceAll(platform, "${"+argName+"}", args[argName]))
				}
			case false:
				arch = parsePlatform(platform)
			}

			match := regexVariableName.Match([]byte(value))
			switch match {
			case true:
				prefix, argName, suffix := extractArgName(value)

				if _, found := args[argName]; found {
					inst := instruction{
						name:          ArgInstruction,
						value:         argName,
						arch:          arch,
						image:         strings.ReplaceAll(value, "${"+argName+"}", args[argName]),
						trimArgPrefix: prefix,
						trimArgSuffix: suffix,
					}
					instructions = append(instructions, inst)
				} else {
					logrus.Debugf("No arg key %q found", argName)
				}

			case false:
				inst := instruction{
					name:  FromInstruction,
					value: value,
					arch:  arch,
					image: value,
				}
				instructions = append(instructions, inst)
			}

		// If we identify an ARG key/value then we store that information to use later
		// with a FROM instruction to rebuild the docker image name + tag
		// So we can be able to update the ARG instruction
		case ArgInstruction:
			lastArgs := searchArgsValue(n)
			// Override old args if the same key is specified multiple time
			for key, value := range lastArgs {
				args[key] = value
			}
		}
		i++
	}
	return instructions, err
}

func searchInstructionFlag(flagName string, flags []string) string {
	instructionPrefix := "--" + flagName + "="
	for _, flag := range flags {
		if strings.HasPrefix(flag, instructionPrefix) {
			return strings.TrimPrefix(flag, instructionPrefix)
		}
	}
	return ""
}

func searchFromValue(n *parser.Node) string {
	for nod := n.Next; nod != nil; nod = nod.Next {
		return nod.Value
	}
	return ""
}

func searchArgsValue(n *parser.Node) map[string]string {
	args := map[string]string{}
	for nod := n.Next; nod != nil; nod = nod.Next {

		switch strings.Contains(nod.Value, "=") {
		case true:
			s := strings.Split(nod.Value, "=")
			args[s[0]] = s[1]
		case false:
			args[nod.Value] = ""

			if nod.Next != nil {
				args[nod.Value] = nod.Next.Value
				nod = nod.Next
			}
		}
	}
	return args
}

func extractArgName(value string) (prefix, argName, suffix string) {
	found := regexVariableName.FindStringSubmatch(value)
	if len(found) > 0 {
		prefix = found[1]
		argName = found[2]
		suffix = found[3]
	}

	return prefix, argName, suffix
}

func parsePlatform(platform string) (arch string) {
	p := strings.Split(platform, "/")

	switch len(p) {
	case 3:
		arch = p[1]
	case 2:
		arch = p[1]
	}

	return arch
}
