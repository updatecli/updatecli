package github

import (
	"bytes"
	"fmt"
	"text/template"
)

// PULLREQUESTBODY is the pull request template used as pull request description
const PULLREQUESTBODY string = `

## Changelog

{{ .Description }}

## Reports

{{ .Report }}

## Remark

This pull request was automatically created using [Updatecli](https://www.updatecli.io).
Please report any issues with this tool [here](https://github.com/updatecli/updatecli/issues/new)
`

// PullRequest contains multiple fields mapped to Github V4 api
type PullRequest struct {
	BaseRefName string
	Body        string
	HeadRefName string
	ID          string
	State       string
	Title       string
	Url         string
}

// SetBody generates the body pull request based on PULLREQUESTBODY
func SetBody(changelog Changelog) (body string, err error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODY))

	body = ""

	buffer := new(bytes.Buffer)

	fmt.Println(changelog)

	err = t.Execute(buffer, changelog)

	if err != nil {
		return "", err
	}

	body = buffer.String()

	return body, err
}
