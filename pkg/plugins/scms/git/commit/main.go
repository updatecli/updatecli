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

	commit, err := c.ParseMessage(raw)

	if err != nil {
		logrus.Errorf(err.Error())
		return "", err
	}

	buffer := new(bytes.Buffer)

	err = t.Execute(buffer, commit)

	commitMessage := buffer.String()

	if err != nil {
		return "", err
	}

	return commitMessage, nil
}

// ParseMessage parses a message then return the commit message title and its body.
// The message title can't be longer than 72 characters
func (c *Commit) ParseMessage(message string) (Commit, error) {
	var commit Commit

	if len(message) == 0 {
		return commit, ErrEmptyCommitMessage
	}

	commit.Type = c.Type
	commit.Scope = c.Scope
	commit.Footers = c.Footers
	commit.HideCredit = c.HideCredit

	lines := strings.Split(message, "\n")

	placeholders := 0
	if len(c.Type) > 0 {
		// Counting colon and space after type
		// This must be updated if commitTpl is changed
		placeholders = placeholders + len(c.Type) + 2
	}
	if len(c.Scope) > 0 {
		// Counting parentheses around the scope
		// This must be updated if commitTpl is changed
		placeholders = placeholders + len(c.Scope) + 2
	}
	maxMessageCharacter := 72 - placeholders

	// If Title based on first line message is too long
	// then split the message title into the body
	if len(lines[0]) > maxMessageCharacter {
		line := lines[0]
		commit.Title = line[0:(maxMessageCharacter-3)] + "..."
		commit.Body = commit.Body + "... " + line[(maxMessageCharacter-3):] + "\n"

	} else {
		commit.Title = lines[0]
	}

	if len(lines) > 1 {
		commit.Body = commit.Body + strings.Join(lines[1:], "\n")

	}

	// Remove trailing \b so we can handle it from the go template
	commit.Body = strings.TrimRight(commit.Body, "\n")

	// Only set title if not already defined
	// so we can specify its value from the updatecli configuration
	// If title is set, then body is also set, even if it's empty
	// If title is not set, but body is, we ignore body, otherwise the commit
	// message will be without context - title is not set, but body is
	if len(c.Title) > 0 {
		commit.Title = c.Title
		commit.Body = c.Body
	}

	return commit, nil
}

// Validate validates "conventional commit" default parameters.
func (c *Commit) Validate() error {
	if len(c.Type) == 0 {
		c.Type = "chore"
	}

	return nil
}
