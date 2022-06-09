package git

import (
	"strings"
	"testing"
)

type DataSet []Data

type Data struct {
	s                     Spec
	expectedDirectoryName string
}

var (
	Dataset DataSet = DataSet{
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Username: "git",
				Password: "password",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Username: "git",
				Password: "password",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Password: "password",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Username: "git",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://@github.com/updatecli/updatecli.git",
				Username: "git",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "ssh://server:project.git",
				Username: "bob",
			}, expectedDirectoryName: "server_project_git",
		},
	}
)

func TestSanitizeDirectoryName(t *testing.T) {
	for _, data := range Dataset {
		got := sanitizeDirectoryName(data.s.URL)

		if strings.Compare(got, data.expectedDirectoryName) != 0 {
			t.Errorf("got sanitize directory name %q, expected %q", got, data.expectedDirectoryName)
		}
	}
}
