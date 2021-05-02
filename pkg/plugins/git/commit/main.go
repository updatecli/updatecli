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
	Version string // Define conventionalCommit version
	Type    string // Define commit type, like chore, fix, etc
	Scope   string // Define commit type scope
	Footers string // Define commit footer
	Title   string // Define commit title
	Body    string // Define commit body
}

// Generate generates the conventional commit
func (c *Commit) Generate(raw string) (string, error) {

	err := c.Validate()

	if err != nil {
		logrus.Errorf(err.Error())
		return "", err
	}

	t := template.Must(template.New("commit").Parse(commitTpl))

	title, body, err := ParseMessage(raw)

	if err != nil {
		logrus.Errorf(err.Error())
		return "", err
	}

	if len(c.Title) == 0 {
		c.Title = title
	}

	if len(c.Body) == 0 {
		c.Body = body
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
func ParseMessage(message string) (title, body string, err error) {
	if len(message) == 0 {
		return "", "", ErrEmptyCommitMessage
	}
	lines := strings.Split(message, "\n")

	// If Title based on first line message is too long
	// then split the message title into the body
	if len(lines[0]) > 72 {
		line := lines[0]
		title = line[0:69] + "..."
		body = body + "... " + line[69:] + "\n"

	} else {
		title = lines[0]
	}

	if len(lines) > 1 {
		body = body + strings.Join(lines[1:], "\n")
	}

	// Remove trailing \b so we can handle it from the go template
	body = strings.TrimRight(body, "\n")

	return title, body, nil
}

// Validate validates "conventional commit" default parameters.
func (c *Commit) Validate() error {
	if len(c.Type) == 0 {
		c.Type = "chore"
	}

	return nil
}
