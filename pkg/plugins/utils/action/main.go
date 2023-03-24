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

<details><summary>Updatecli options</summary>
Most of Updatecli configuration is done via Updatecli manifest.
<ul>
<li>If you close this pullrequest, Updatecli will automatically reopen it, next it runs.</li>
<li>If you close this pullrequest, and delete the base branch, Updatecli will automatically recreate it, erasing all previous commits made.</li>
</ul>
</details>

---

Action triggered automatically by [Updatecli](https://www.updatecli.io).

Feel free to report any issues at [github.com/updatecli/updatecli](https://github.com/updatecli/updatecli/issues/).
If you find this tool useful, do not hesitate to star our GitHub repository [github.com/updatecli/updatecli](https://github.com/updatecli/updatecli/stargazers) as a sign of appreciation.
Or tell us directly on our [https://matrix.to/#/#Updatecli_community:gitter.im](chat)
`

// GeneratePullRequestBody generates the Pull Request's body based on PULLREQUESTBODY
func GeneratePullRequestBody(Description, Report string) (string, error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODYTEMPLATE))

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
