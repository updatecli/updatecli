package commit

import (
	"strings"
	"testing"
)

type DataSet []Data

type Data struct {
	Message        string
	ExpectedOutput string
	ExpectedError  error
	ExpectedTitle  string
	ExpectedBody   string
	Commit         Commit
}

var (
	dataset = DataSet{
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore(deps): Bump updatecli version",
			ExpectedError:  nil,
			ExpectedBody:   "",
			ExpectedTitle:  "Bump updatecli version",
			Commit: Commit{
				Type:        "chore",
				Scope:       "deps",
				HideCredits: true,
			},
		},
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore: Bump updatecli version\nMade with ❤️️ by updatecli",
			ExpectedError:  nil,
			ExpectedBody:   "",
			ExpectedTitle:  "Bump updatecli version",
			Commit:         Commit{},
		},
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore: Bump updatecli version",
			ExpectedError:  nil,
			ExpectedBody:   "",
			ExpectedTitle:  "Bump updatecli version",
			Commit: Commit{
				Type:        "chore",
				HideCredits: true,
			},
		},
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore: Bump updatecli version\n\nBREAKING CHANGE",
			ExpectedError:  nil,
			ExpectedBody:   "",
			ExpectedTitle:  "Bump updatecli version",
			Commit: Commit{
				Type:        "chore",
				Footers:     "BREAKING CHANGE",
				HideCredits: true,
			},
		},
		{
			Message:        strings.Repeat("a", 75),
			ExpectedOutput: "chore: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...\n\n... aaaaaaaaaaa\n\nBREAKING CHANGE",
			ExpectedError:  nil,
			ExpectedBody:   "... aaaaaaaaaaa",
			ExpectedTitle:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...",
			Commit: Commit{
				Type:        "chore",
				Footers:     "BREAKING CHANGE",
				HideCredits: true,
			},
		},
		{
			Message:        strings.Repeat("a", 75),
			ExpectedOutput: "chore: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...\n\n... aaaaaaaaaaa",
			ExpectedError:  nil,
			ExpectedBody:   "... aaaaaaaaaaa",
			ExpectedTitle:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...",
			Commit: Commit{
				Type:        "chore",
				HideCredits: true,
			},
		},
		{
			Message:        "",
			ExpectedOutput: "",
			ExpectedError:  ErrEmptyCommitMessage,
			ExpectedBody:   "",
			ExpectedTitle:  "",
			Commit: Commit{
				Type:        "chore",
				HideCredits: true,
			},
		},
	}
)

func TestCommit(t *testing.T) {

	for id, data := range dataset {
		got, err := data.Commit.Generate(data.Message)
		if err != nil && data.ExpectedError != nil {
			if strings.Compare(err.Error(), data.ExpectedError.Error()) != 0 {
				t.Errorf("Wrong commit %d err:\n\tExpected:\t\t%v\n\tGot:\t\t%v\n", id, data.ExpectedError, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q for commit #%d", err, id)
		}
		if strings.Compare(data.ExpectedOutput, got) != 0 {
			t.Errorf("Wrong Commit Message %d:\n\tGot:\t\t%q\n\tExpected:\t%q",
				id,
				got,
				data.ExpectedOutput)
		}
	}
}

func TestParseMessage(t *testing.T) {

	for id, data := range dataset {
		err := data.Commit.ParseMessage(data.Message)

		if err != nil && data.ExpectedError != nil {
			if strings.Compare(err.Error(), data.ExpectedError.Error()) != 0 {
				t.Errorf("Wrong commit %d err:\n\tExpected:\t\t%v\n\tGot:\t\t%v\n", id, data.ExpectedError, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q for commit #%d", err, id)
		}

		if strings.Compare(data.ExpectedTitle, data.Commit.Title) != 0 {
			t.Errorf("Wrong Commit Title %d:\n\tGot:\t\t%q\n\tExpected:\t%q",
				id,
				data.Commit.Title,
				data.ExpectedTitle)
		}
		if strings.Compare(data.ExpectedBody, data.Commit.Body) != 0 {
			t.Errorf("Wrong Commit Body %d:\n\tGot:\t\t%q\n\tExpected:\t%q",
				id,
				data.Commit.Body,
				data.ExpectedBody)
		}
	}
}
