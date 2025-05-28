package reports

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Actions struct {
	Actions []Action `xml:"action"`
}

func (a *Actions) Merge(sourceActions *Actions) {
	var c, d Actions
	useDetailsFromSourceAction := true
	switch len(a.Actions) > len(sourceActions.Actions) {
	case true:
		c = *a
		d = *sourceActions
		useDetailsFromSourceAction = false
	case false:
		d = *a
		c = *sourceActions
	}

	for i := range d.Actions {
		pipelineFound := false
	out:
		for j := range c.Actions {
			if d.Actions[i].ID == c.Actions[j].ID {
				pipelineFound = true
				c.Actions[j].Merge(&d.Actions[i], useDetailsFromSourceAction)
				break out
			}
		}
		if !pipelineFound {
			c.Actions = append(c.Actions, d.Actions[i])
		}
	}
	*a = c
}

// String show an action report formatted as a string
func (a *Actions) String() string {
	a.sort()

	output, err := xml.MarshalIndent(a, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return string(output[:])
}

func MergeFromString(old, new string) string {
	var oldReport Actions
	var newReport Actions

	if old == "" && new != "" {
		return new
	}

	if old != "" && new == "" {
		return old
	}

	err := unmarshal([]byte(old), &oldReport)
	if err != nil {
		logrus.Errorf("failed parsing old report: %s", err)
		// Return the new report by default if something went wrong
		return new
	}

	err = unmarshal([]byte(new), &newReport)
	if err != nil {
		logrus.Errorf("failed parsing new report: %s", err)
		// Return the new report by default
		return new
	}

	newReport.Merge(&oldReport)

	return newReport.String()
}

func (a *Actions) sort() {
	actions := *a
	sort.Slice(
		actions.Actions,
		func(i, j int) bool {
			return actions.Actions[i].ID < actions.Actions[j].ID
		})
}

// unmarshal parses the htmlReport string and return a struct
func unmarshal(input []byte, a *Actions) (err error) {
	if err := xml.Unmarshal(input, a); err != nil {
		return err
	}

	b := *a
	for i := range b.Actions {
		b.Actions[i].sort()
	}
	return nil
}

func markdownToActions(input string, actions *Actions) error {
	source := []byte(input)

	actionCount := -1
	actionTargetCount := -1
	actionTargetChangeLogCount := -1

	md := goldmark.New()
	node := md.Parser().Parse(text.NewReader(source))

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			title := string(n.Lines().Value(source))

			switch n.Level {
			case 1:
				action := Action{
					PipelineTitle: title,
				}
				actions.Actions = append(actions.Actions, action)
				actionCount++
				actionTargetCount = -1
				actionTargetChangeLogCount = -1
			case 2:
				actionTarget := ActionTarget{
					Title: title,
				}
				actions.Actions[actionCount].Targets = append(actions.Actions[actionCount].Targets, actionTarget)
				actionTargetCount++
				actionTargetChangeLogCount = -1
			case 3:
				actionTargetChangelog := ActionTargetChangelog{
					Title: title,
				}
				actions.Actions[actionCount].Targets[actionTargetCount].Changelogs = append(
					actions.Actions[actionCount].Targets[actionTargetCount].Changelogs,
					actionTargetChangelog)
				actionTargetChangeLogCount++
			}
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph:
			contents := n.Lines().Value(source)
			if child, ok := n.FirstChild().(*ast.Text); ok {
				switch string(child.Value(source)) {
				case "Pipeline ID: ":
					if sibling, ok := child.NextSibling().(*ast.CodeSpan); ok {
						if codespanValue, ok := sibling.FirstChild().(*ast.Text); ok {
							actions.Actions[actionCount].ID = string(codespanValue.Value(source))
						}
					}
					return ast.WalkSkipChildren, nil
				case "Target ID: ":
					if sibling, ok := child.NextSibling().(*ast.CodeSpan); ok {
						if codespanValue, ok := sibling.FirstChild().(*ast.Text); ok {
							actions.Actions[actionCount].Targets[actionTargetCount].ID = string(codespanValue.Value(source))
						}
					}
					return ast.WalkSkipChildren, nil
				case "Pipeline URL: ":
					if sibling, ok := child.NextSibling().(*ast.Link); ok {
						var name string
						if codespanValue, ok := sibling.FirstChild().(*ast.Text); ok {
							name = string(codespanValue.Value(source))
						}
						actions.Actions[actionCount].PipelineURL = &PipelineURL{
							Name: name,
							URL:  string(sibling.Destination),
						}
					}
					return ast.WalkSkipChildren, nil
				}
			}
			if actionCount > -1 && actionTargetCount > -1 {
				actions.Actions[actionCount].Targets[actionTargetCount].Description = string(contents)
				return ast.WalkSkipChildren, nil
			}
		case *ast.FencedCodeBlock:
			if actionTargetChangeLogCount > -1 {
				contents := n.Lines().Value(source)
				actions.Actions[actionCount].Targets[actionTargetCount].Changelogs[actionTargetChangeLogCount].Description = strings.TrimSpace(string(contents))
				return ast.WalkSkipChildren, nil
			}
		case *ast.List:
			if actionCount > -1 && actionTargetCount > -1 {
				var buf bytes.Buffer
				for children := n.FirstChild(); children != nil; children = children.NextSibling() {
					if listValue, ok := children.FirstChild().(*ast.TextBlock); ok {
						buf.Write([]byte("\n* "))
						buf.Write(listValue.Lines().Value(source))
					}
				}
				actions.Actions[actionCount].Targets[actionTargetCount].Description += "\n"
				actions.Actions[actionCount].Targets[actionTargetCount].Description += buf.String()
				return ast.WalkSkipChildren, nil
			}
		case *ast.ThematicBreak:
			// start of footer stop processing
			return ast.WalkStop, nil
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func MergeFromMarkdown(old, new string) (string, error) {
	var oldReport Actions
	var newReport Actions

	if old == "" && new != "" {
		return new, nil
	}

	if old != "" && new == "" {
		return old, nil
	}

	if err := markdownToActions(old, &oldReport); err != nil {
		return "", err
	}

	if err := markdownToActions(new, &newReport); err != nil {
		return "", err
	}

	newReport.Merge(&oldReport)
	newReport.sort()

	var report string

	for i, action := range newReport.Actions {
		if i > 0 {
			report += "\n\n"
		}
		r := action.ToActionsMarkdownString()
		report += r
	}

	return report, nil
}
