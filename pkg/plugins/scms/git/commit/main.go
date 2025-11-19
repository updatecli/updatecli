package commit

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

const (
	// commitTpl is the template used to generate the commit message
	commitTpl string = "{{ .Type}}{{if .Scope}}({{.Scope}}){{ end }}: {{ .Title }}" +
		"{{ if .Body }}\n\n{{ .Body }}{{ end }}" +
		"{{ if not .HideCredit }}\n\nMade with ❤️️ by updatecli{{end}}" +
		"{{ if .Footers }}\n\n{{ .Footers }}{{ end }}"
)

var (
	// ErrEmptyCommitMessage is returned when the commit message is empty
	ErrEmptyCommitMessage error = errors.New("error: empty commit message")
)

/*
Commit contains conventional commit information
More information on what is conventional commits
-> https://www.conventionalcommits.org/en/v1.0.0/#summary
*/
type Commit struct {
	//
	//  type defines the type of commit message such as "chore", "fix", "feat", etc. as
	//  defined by the conventional commit specification. More information on
	//  -> https://www.conventionalcommits.org/en/
	//
	//  default:
	//    * chore
	//
	Type string `yaml:",omitempty"`
	//
	//  scope defines the scope of the commit message as defined by the
	//  conventional commit specification. More information on
	//  -> https://www.conventionalcommits.org/en/
	//
	//  default:
	//    none
	//
	Scope string `yaml:",omitempty"`
	//
	//  footers defines the footer of the commit message as defined by the
	//  conventional commit specification. More information on
	//  -> https://www.conventionalcommits.org/en/
	//
	//  default:
	//    none
	//
	Footers string `yaml:",omitempty"`
	//
	//  title defines the title of the commit message as defined by the
	//  conventional commit specification. More information on
	//  -> https://www.conventionalcommits.org/en/
	//
	//  default:
	//    default is set to the target name or the target short description
	//    if the name is not defined.
	//
	Title string `yaml:",omitempty"`
	//
	//  body defines the commit body of the commit message as defined by the
	//  conventional commit specification. More information on
	//  -> https://www.conventionalcommits.org/en/
	//
	//  default:
	//    none
	//
	Body string `yaml:",omitempty"`
	//
	//  hideCredit defines if updatecli credits should be displayed inside commit message body
	//
	//  please consider sponsoring the Updatecli project if you want to disable credits.
	//  -> https://github.com/updatecli/updatecli
	//
	//  default:
	//  	false
	//
	HideCredit bool `yaml:",omitempty"`
	//  squash defines if the commit should be squashed
	//
	//  default:
	//  	false
	//
	//  important:
	//   if squash is set to true, then it's highly recommended to set the commit title and body
	//   to a meaningful value as all other commit information will be lost during the squash operation.
	//
	//   if title and body are not set, then the commit title/message will be generated based on the most recent commit
	//   message of the squashed commits.
	//
	Squash *bool `yaml:",omitempty"`
}

/*
Generate generates the conventional commit
*/
func (c *Commit) Generate(raw string) (string, error) {

	err := c.Validate()

	if err != nil {
		logrus.Errorln(err.Error())
		return "", err
	}

	t := template.Must(template.New("commit").Parse(commitTpl))

	commit, err := c.ParseMessage(raw)

	if err != nil {
		logrus.Errorln(err.Error())
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

/*
ParseMessage parses a message then return the commit message title and its body.
The message title can't be longer than 72 characters
*/
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

/*
Validate validates "conventional commit" default parameters.
*/
func (c *Commit) Validate() error {
	if len(c.Type) == 0 {
		c.Type = "chore"
	}

	return nil
}

// IsSquash returns true if the commit should be squashed
func (c *Commit) IsSquash() bool {
	if c.Squash == nil {
		return false
	}
	return *c.Squash
}
