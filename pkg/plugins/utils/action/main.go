package action

import (
	"bytes"
	"text/template"

	"github.com/updatecli/updatecli/pkg/plugins/utils/truncate"
)

// PULLREQUESTBODYTEMPLATE is the template used as a Pull Request description
// Please note that triple backticks are concatenated with the literals, as they cannot be escaped
const PULLREQUESTBODYTEMPLATE = `
{{ if .PreDescription }}
{{ .PreDescription }}

---

{{ end }}

{{ .Report }}

---

<table>
  <tr>
    <td width="77">
      <img src="https://www.updatecli.io/images/updatecli.png" alt="Updatecli logo" width="50" height="50" />
    </td>
    <td>
      <p>
        Created automatically by <a href="https://www.updatecli.io/">Updatecli</a>
      </p>
      <details><summary>Options:</summary>
        <br />
        <p>Most of Updatecli configuration is done via <a href="https://www.updatecli.io/docs/prologue/quick-start/">its manifest(s)</a>.</p>
        <ul>
          <li>If you close this pull request, Updatecli will automatically reopen it, the next time it runs.</li>
          <li>If you close this pull request and delete the base branch, Updatecli will automatically recreate it, erasing all previous commits made.</li>
        </ul>
        <p>
          Feel free to report any issues at <a href="https://github.com/updatecli/updatecli/issues">github.com/updatecli/updatecli</a>.<br />
          If you find this tool useful, do not hesitate to star <a href="https://github.com/updatecli/updatecli/stargazers">our GitHub repository</a> as a sign of appreciation, and/or to tell us directly on our <a href="https://matrix.to/#/#Updatecli_community:gitter.im">chat</a>!
        </p>
      </details>
    </td>
  </tr>
</table>

`

const PULLREQUESTBODYTEMPLATEMARKDOWN = `
{{ if .PreDescription }}
{{ .PreDescription }}

---

{{ end }}

{{ .Report }}

---

Created automatically by [Updatecli](https://www.updatecli.io/)

Most of Updatecli configuration is done via [its manifest(s)](https://www.updatecli.io/docs/prologue/quick-start/).

* If you close this pull request, Updatecli will automatically reopen it, the next time it runs.
* If you close this pull request and delete the base branch, Updatecli will automatically recreate it, erasing all previous commits made.

Feel free to report any issues at [github.com/updatecli/updatecli](https://github.com/updatecli/updatecli/issues).
If you find this tool useful, do not hesitate to star [our GitHub repository](https://github.com/updatecli/updatecli/stargazers) as a sign of appreciation, and/or to tell us directly on our [chat](https://matrix.to/#/#Updatecli_community:gitter.im)!
`

// GeneratePullRequestBody generates the Pull Request's body based on PULLREQUESTBODYTEMPLATE
func GeneratePullRequestBody(Description, Report string) (string, error) {
	return generatePullRequestBodyFromTemplate(Description, Report, PULLREQUESTBODYTEMPLATE)
}

// GeneratePullRequestBodyMarkdown generates the Pull Request's body based on PULLREQUESTBODYTEMPLATEMARKDOWN
func GeneratePullRequestBodyMarkdown(Description, Report string) (string, error) {
	return generatePullRequestBodyFromTemplate(Description, Report, PULLREQUESTBODYTEMPLATEMARKDOWN)
}

// generatePullRequestBodyFromTemplate generates the Pull Request's from provided template
func generatePullRequestBodyFromTemplate(Description, Report, Template string) (string, error) {
	t := template.Must(template.New("pullRequest").Parse(Template))

	buffer := new(bytes.Buffer)

	type params struct {
		Report         string
		PreDescription string
	}

	err := t.Execute(buffer, params{
		PreDescription: Description,
		Report:         Report,
	})
	if err != nil {
		return "", err
	}

	// GitHub Issues/PRs messages have a max size limit on the
	// message body payload.
	// `body is too long (maximum is 65536 characters)`.
	// To avoid that, we ensure to cap the message to 65k chars.
	const MAX_CHARACTERS_PER_MESSAGE = 65000

	return truncate.String(buffer.String(), MAX_CHARACTERS_PER_MESSAGE), nil
}
