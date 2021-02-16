package mobyparser

import (
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Marshal takes a dockerfile node as parameter,
// parse its AST graph then return line by line as a string
// Remark: Dockerfile Comments are not yet supported and
// therefor deleted in the process
func Marshal(data *parser.Result, document *string) (err error) {

	arguments := ""
	tmp := []string{}

	for _, node := range data.AST.Children {

		instruction := strings.ToUpper(node.Value)
		tab := strings.Repeat(" ", len(node.Value)+1)

		switch instruction {
		case "FROM":
			arguments = DefaultForm(node)
		case "LABEL":
			arguments = KeyValueForm(node, tab)
		case "MAINTAINER":
			arguments = DefaultForm(node)
		case "EXPOSE":
			arguments = DefaultForm(node)
		case "ADD":
			arguments = DefaultForm(node)
		case "ONBUILD":
			for _, n := range node.Next.Children {
				arguments = strings.ToUpper(n.Value) + " " + DefaultForm(n)
			}
		case "STOPSIGNAL":
			arguments = DefaultForm(node)
		case "HEALTHCHECK":
			arguments = DefaultForm(node)
		case "ARG":
			arguments = KeyValueForm(node, tab)
		case "COPY":
			arguments = DefaultForm(node)
		case "ENV":
			arguments = KeyValueForm(node, tab)
		case "RUN":
			arguments = ShellForm(node)
			//arguments = ExecForm(node)
		case "CMD":
			arguments = ExecForm(node)
			//arguments = ShellForm(node)
		case "ENTRYPOINT":
			arguments = ExecForm(node)
			//arguments = ShellForm(node)
		case "SHELL":
			arguments = ExecForm(node)
			//arguments = ShellForm(node)
		case "VOLUME":
			//arguments = ExecForm(node)
			arguments = DefaultForm(node)
		case "USER":
			arguments = DefaultForm(node)

		case "WORKDIR":
			arguments = DefaultForm(node)

		default:
			return fmt.Errorf("Instruction %s not supported", instruction)
		}

		if len(arguments) > 0 {
			tmp = append(tmp, fmt.Sprintf("%s %s", instruction, arguments))
		} else {
			tmp = append(tmp, instruction)
		}

	}

	*document = strings.Join(tmp, "\n")

	return err
}

// DefaultForm format is the default instruction line
func DefaultForm(node *parser.Node) (arguments string) {

	for n := node.Next; n != nil; n = n.Next {
		if arguments != "" {
			arguments = fmt.Sprintf("%s %s", arguments, n.Value)
		} else {
			arguments = n.Value
		}
	}
	if len(node.Flags) > 0 {
		arguments = fmt.Sprintf("%s %s", strings.Join(node.Flags, " "), arguments)
	}
	return arguments + "\n"
}

// ExecForm format instruction arguments to an exec form, like ["/bin/bash"]
func ExecForm(node *parser.Node) (arguments string) {
	tmp := []string{}

	for n := node.Next; n != nil; n = n.Next {
		value := n.Value
		if strings.HasPrefix(n.Value, `"`) && strings.HasSuffix(n.Value, `"`) {
			value = strings.TrimPrefix(value, `"`)
			value = strings.TrimSuffix(value, `"`)
		}
		tmp = append(tmp, `"`+strings.ReplaceAll(value, `"`, `\"`)+`"`)
	}

	arguments = `[ ` + strings.Join(tmp, `,`) + ` ]`

	if len(node.Flags) > 0 {
		arguments = fmt.Sprintf("%s %s", strings.Join(node.Flags, " "), arguments)
	}

	return arguments + "\n"
}

// KeyValueForm format instruction arguments and add '=' between a key and a value,
// like `LABEL key=value`
func KeyValueForm(node *parser.Node, tab string) (arguments string) {

	tmp := []string{}

	for n := node.Next; n != nil; n = n.Next {
		tmp = append(tmp, n.Value)
	}
	separator := ""
	endline := ""
	for i, raw := range tmp {
		if (i % 2) == 0 {
			separator = ""
			endline = ""
		} else if (i % 2) != 0 {
			separator = "="
			if i < (len(tmp) - 1) {
				endline = fmt.Sprintf("\\\n%s", tab)
			}
		}
		arguments = fmt.Sprintf("%s%s%s%s",
			arguments,
			separator,
			raw,
			endline)
	}

	if len(node.Flags) > 0 {
		arguments = fmt.Sprintf("%s %s", strings.Join(node.Flags, " "), arguments)
	}

	return arguments + "\n"
}

// ShellForm format a line instruction containing shell commands,
// like RUN "echo true && echo false"
func ShellForm(node *parser.Node) (arguments string) {
	separator := ""
	for n := node.Next; n != nil; n = n.Next {
		replacer := "&& \\\n"
		raw := strings.ReplaceAll(n.Value, "&& ", replacer)
		arguments = fmt.Sprintf("%s%s%s", arguments, separator, raw)
		separator = " "
	}

	if len(node.Flags) > 0 {
		arguments = fmt.Sprintf("%s %s", strings.Join(node.Flags, " "), arguments)
	}

	return arguments + "\n"
}
