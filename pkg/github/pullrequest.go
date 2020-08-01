package github

import (
	"bytes"
	"text/template"
)

// PULLREQUESTBODY contains the Github Pull request body
// TODO Add changelog
// TODO Explain source
// TODO Add conditions if exist
const PULLREQUESTBODY string = `
{{ .Changelog }}

This pull request was automatically created using [olblak/updatecli](https://github.com/olblak/updatecli).
Based on a source rule, it checks if yaml value can be update to the latest version
Please carefully review yaml modification as it also reformat it.
Please report any issues with this tool [here](https://github.com/olblak/updatecli/issues/new)
`

// PullRequest contains multiple fields mapped on Github V4 api...
type PullRequest struct {
	BaseRefName string
	Body        string
	HeadRefName string
	ID          string
	State       string
	Title       string
	Url         string
}

// SetBody generate the body pull request based on the template PULLREQUESTBODY
func SetBody(changelog string) (body string, err error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODY))

	body = ""

	buffer := new(bytes.Buffer)

	type params struct {
		Changelog string
	}

	err = t.Execute(buffer, params{
		Changelog: changelog})

	if err != nil {
		return "", err
	}

	body = buffer.String()

	return body, err
}
