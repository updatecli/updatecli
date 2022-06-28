package commit

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

const (
	commitTpl string = "{{ .Type}}{{if .Scope}}({{.Scope}}){{ end }}: {{ .Title }}" +
		"{{ if .Body }}\n\n{{ .Body }}{{ end }}" +
		"{{ if not .HideCredit }}\n\nMade with ❤️️ by updatecli{{end}}" +
		"{{ if .Footers }}\n\n{{ .Footers }}{{ end }}"
)

var (
	// ErrEmptyCommitMessage is returned when the commit message is empty
	ErrEmptyCommitMessage error = errors.New("error: empty commmit messager")
)

// Commit contains conventional commit information
// More information on what is conventional commits
// -> https://www.conventionalcommits.org/en/v1.0.0/#summary
type Commit struct {
	// Define commit type, like chore, fix, etc
	Type string `yaml:",omitempty"`
	// Define commit type scope
	Scope string `yaml:",omitempty"`
	// Define commit footer
	Footers string `yaml:",omitempty"`
	// Define commit title
	Title string `yaml:",omitempty"`
	// Define commit body
	Body string `yaml:",omitempty"`
	// Display updatecli credits inside commit message body
	HideCredit bool `yaml:",omitempty"`
}

// Generate generates the conventional commit
func (c *Commit) Generate(raw string) (string, error) {

	err := c.Validate()

	if err != nil {
		logrus.Errorf(err.Error())
		return "", err
	}

	t := template.Must(template.New("commit").Parse(commitTpl))

	err = c.ParseMessage(raw)

	if err != nil {
		logrus.Errorf(err.Error())
		return "", err
	}

	buffer := new(bytes.Buffer)

	err = t.Execute(buffer, c)

	commitMessage := buffer.String()

	if err != nil {
		return "", err
	}

	return commitMessage, nil
}

// ParseMessage parses a message then return the commit message title and its body.
// The message title can't be longer than 72 characters
func (c *Commit) ParseMessage(message string) (err error) {
	if len(message) == 0 {
		return ErrEmptyCommitMessage
	}
	lines := strings.Split(message, "\n")

	maxMessageCharacter := 72 - (len(c.Type) + len(c.Scope))
	title := ""
	body := ""

	// If Title based on first line message is too long
	// then split the message title into the body
	if len(lines[0]) > maxMessageCharacter {
		line := lines[0]
		title = line[0:(maxMessageCharacter-3)] + "..."
		body = body + "... " + line[(maxMessageCharacter-3):] + "\n"

	} else {
		title = lines[0]
	}

	if len(lines) > 1 {
		body = body + strings.Join(lines[1:], "\n")
	}

	// Remove trailing \b so we can handle it from the go template
	body = strings.TrimRight(body, "\n")

	// Only set title if not already defined
	// so we can specify its value from the updatecli configuration
	if len(c.Title) == 0 {
		c.Title = title
	}
	// Only set body if not already defined
	// so we can specify its value from the updatecli configuration
	if len(c.Body) == 0 {
		c.Body = body
	}

	return nil
}

// Validate validates "conventional commit" default parameters.
func (c *Commit) Validate() error {
	if len(c.Type) == 0 {
		c.Type = "chore"
	}

	return nil
}
