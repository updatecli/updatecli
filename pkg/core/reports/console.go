package reports

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// stageItem holds what the console summary shows for a source, condition or target
type stageItem struct {
	Result    string
	Name      string
	ScmURL    string
	ScmBranch string
}

// render returns the console summary of every report, without leading or trailing new lines
func (r Reports) render() string {
	blocks := make([]string, 0, len(r))
	for i := range r {
		blocks = append(blocks, r[i].renderConsole())
	}
	return strings.Join(blocks, "\n")
}

// renderConsole returns the console summary of a single report, ending with a new line
func (r *Report) renderConsole() string {
	b := &strings.Builder{}

	fmt.Fprintf(b, "%s %s:\n", r.Result, r.Name)

	if r.Err != "" {
		if r.PipelineID != "" {
			fmt.Fprintf(b, "%s\n", r.PipelineID)
		}
		fmt.Fprintf(b, "\tError: %s\n", r.Err)
		return b.String()
	}

	writeStage(b, "Source", stageItems(r.Sources, func(s *result.Source) stageItem {
		return stageItem{Result: s.Result, Name: s.Name, ScmURL: s.Scm.URL, ScmBranch: s.Scm.Branch.Target}
	}))
	writeStage(b, "Condition", stageItems(r.Conditions, func(c *result.Condition) stageItem {
		return stageItem{Result: c.Result, Name: c.Name, ScmURL: c.Scm.URL, ScmBranch: c.Scm.Branch.Target}
	}))
	writeStage(b, "Target", stageItems(r.Targets, func(t *result.Target) stageItem {
		return stageItem{Result: t.Result, Name: t.Name, ScmURL: t.Scm.URL, ScmBranch: t.Scm.Branch.Target}
	}))
	writeActions(b, r.Actions)

	if r.ReportURL != "" {
		fmt.Fprintf(b, "\n\t=> Report available on %s\n", r.ReportURL)
	}

	return b.String()
}

// stageItems converts a stage map into items sorted by ID
func stageItems[T any](m map[string]*T, convert func(*T) stageItem) map[string]stageItem {
	items := make(map[string]stageItem, len(m))
	for id, v := range m {
		if v != nil {
			items[id] = convert(v)
		}
	}
	return items
}

// writeStage writes a stage section such as "Source:", followed by one line per item
func writeStage(b *strings.Builder, header string, items map[string]stageItem) {
	if len(items) == 0 {
		return
	}

	fmt.Fprintf(b, "\t%s:\n", header)
	for _, id := range slices.Sorted(maps.Keys(items)) {
		item := items[id]
		fmt.Fprintf(b, "\t\t%s [%s] %s\n", item.Result, id, item.Name)
		if item.ScmURL != "" {
			fmt.Fprintf(b, "\t\t      * Repository: %s (branch: %s)\n", item.ScmURL, item.ScmBranch)
		}
	}
}

// writeActions writes the "Action:" section, with the link of each published action
func writeActions(b *strings.Builder, actions map[string]*Action) {
	if len(actions) == 0 {
		return
	}

	fmt.Fprintf(b, "\tAction:\n")
	for _, id := range slices.Sorted(maps.Keys(actions)) {
		action := actions[id]
		if action == nil {
			continue
		}

		title := action.Title
		if title == "" {
			title = action.PipelineTitle
		}
		fmt.Fprintf(b, "\t\t[%s] %s\n", id, title)

		switch {
		case action.Link != "":
			fmt.Fprintf(b, "\t\t      * Link: %s\n", action.Link)
		case action.DryRun:
			fmt.Fprintf(b, "\t\t      * Not published (dry run)\n")
		}
	}
}
