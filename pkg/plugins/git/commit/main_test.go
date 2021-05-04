package commit

import (
	"strings"
	"testing"
)

type DataSet []Data
type MessageDataSet []MessageData

type Data struct {
	Message        string
	ExpectedOutput string
	ExpectedError  error
	Commit         Commit
}

type MessageData struct {
	Message       string
	ExpectedTitle string
	ExpectedBody  string
	ExpectedError error
}

var (
	dataset = DataSet{
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore(deps): Bump updatecli version",
			ExpectedError:  nil,
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
			Commit:         Commit{},
		},
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore: Bump updatecli version",
			ExpectedError:  nil,
			Commit: Commit{
				Type:        "chore",
				HideCredits: true,
			},
		},
		{
			Message:        "Bump updatecli version",
			ExpectedOutput: "chore: Bump updatecli version\n\nBREAKING CHANGE",
			ExpectedError:  nil,
			Commit: Commit{
				Type:        "chore",
				Footers:     "BREAKING CHANGE",
				HideCredits: true,
			},
		},
		{
			Message:        strings.Repeat("a", 75),
			ExpectedOutput: "chore: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...\n\n... aaaaaa\n\nBREAKING CHANGE",
			ExpectedError:  nil,
			Commit: Commit{
				Type:        "chore",
				Footers:     "BREAKING CHANGE",
				HideCredits: true,
			},
		},
		{
			Message:        strings.Repeat("a", 75),
			ExpectedOutput: "chore: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...\n\n... aaaaaa",
			ExpectedError:  nil,
			Commit: Commit{
				Type:        "chore",
				HideCredits: true,
			},
		},
		{
			Message:        "",
			ExpectedOutput: "",
			ExpectedError:  ErrEmptyCommitMessage,
			Commit: Commit{
				Type:        "chore",
				HideCredits: true,
			},
		},
	}
	messagedataset = MessageDataSet{
		{
			Message:       "Bump updatecli version",
			ExpectedTitle: "Bump updatecli version",
			ExpectedBody:  "",
			ExpectedError: nil,
		},
		{
			Message:       strings.Repeat("a", 72),
			ExpectedTitle: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ExpectedBody:  "",
			ExpectedError: nil,
		},
		{
			Message:       strings.Repeat("a", 75),
			ExpectedTitle: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...",
			ExpectedBody:  "... aaaaaa",
			ExpectedError: nil,
		},
		{
			Message:       "Title\nBody",
			ExpectedTitle: "Title",
			ExpectedBody:  "Body",
			ExpectedError: nil,
		},
		{
			Message:       "",
			ExpectedTitle: "",
			ExpectedBody:  "",
			ExpectedError: ErrEmptyCommitMessage,
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

	for id, data := range messagedataset {
		gotTitle, gotBody, err := ParseMessage(data.Message)

		if err != nil && data.ExpectedError != nil {
			if strings.Compare(err.Error(), data.ExpectedError.Error()) != 0 {
				t.Errorf("Wrong commit %d err:\n\tExpected:\t\t%v\n\tGot:\t\t%v\n", id, data.ExpectedError, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q for commit #%d", err, id)
		}

		if strings.Compare(data.ExpectedTitle, gotTitle) != 0 {
			t.Errorf("Wrong Commit Title %d:\n\tGot:\t\t%q\n\tExpected:\t%q",
				id,
				gotTitle,
				data.ExpectedTitle)
		}
		if strings.Compare(data.ExpectedBody, gotBody) != 0 {
			t.Errorf("Wrong Commit Body %d:\n\tGot:\t\t%q\n\tExpected:\t%q",
				id,
				gotBody,
				data.ExpectedBody)
		}
	}
}
