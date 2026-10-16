package log

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// kind tells how a log line is rendered.
type kind int

const (
	kindMessage kind = iota
	kindTitle
	kindPipelineTitle
	kindPipelineResult
	kindHeading
	kindSection
	kindSubheading
)

// line is a log line, rendered as plain text or styled for a color terminal.
type line struct {
	kind  kind
	level Level
	msg   string
	// category is the category of a section, such as "source" or "target"
	category string
	// id, dryRun, position, total, status and duration describe a pipeline
	id       string
	dryRun   bool
	position int
	total    int
	status   Status
	duration time.Duration
}

// maxRuleWidth is the maximum width of the rule under a step title.
const maxRuleWidth = 100

var (
	styleOnce sync.Once
	styled    bool

	debugStyle    lipgloss.Style
	warningStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	ruleWidth     int
	titleStyle    lipgloss.Style
	bandStyle     lipgloss.Style
	bandIDStyle   lipgloss.Style
	detailStyle   lipgloss.Style
	statusStyles  map[Status]lipgloss.Style
	headingStyle  lipgloss.Style
	categoryStyle lipgloss.Style
	nameStyle     lipgloss.Style
	ruleStyle     lipgloss.Style
)

// styledTerminal reports whether os.Stdout is a terminal supporting colors,
// honoring NO_COLOR and CLICOLOR_FORCE.
func styledTerminal() bool {
	styleOnce.Do(func() {
		r := lipgloss.NewRenderer(os.Stdout)
		styled = r.ColorProfile() != termenv.Ascii
		initStyles(r)

		ruleWidth = maxRuleWidth
		if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 && width < maxRuleWidth {
			ruleWidth = width
		}
	})
	return styled
}

func initStyles(r *lipgloss.Renderer) {
	// The basic ANSI colors follow the terminal theme, light or dark
	debugStyle = r.NewStyle().Foreground(lipgloss.Color("6")).Faint(true)
	warningStyle = r.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	errorStyle = r.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	titleStyle = r.NewStyle().Faint(true)
	bandStyle = r.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	bandIDStyle = r.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("4"))
	detailStyle = r.NewStyle().Faint(true)
	statusStyles = map[Status]lipgloss.Style{
		StatusUnchanged: r.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
		StatusChanged:   r.NewStyle().Foreground(lipgloss.Color("3")).Bold(true),
		StatusFailed:    r.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		StatusSkipped:   r.NewStyle().Faint(true),
	}
	headingStyle = r.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	categoryStyle = r.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	nameStyle = r.NewStyle().Bold(true)
	ruleStyle = r.NewStyle().Faint(true)
}

// plain renders the line as plain text, as stored in reports.
func (l line) plain() string {
	switch l.kind {
	case kindTitle:
		return "\n\n== " + strings.ToTitle(l.msg) + " ==\n\n"
	case kindPipelineTitle:
		title := strings.ToTitle(l.msg)
		if counter := l.counter(); counter != "" {
			title = "[" + counter + "] " + title
		}
		border := strings.Repeat("#", lipgloss.Width(title)+4)
		out := "\n" + border + "\n# " + title + " #\n" + border + "\n\n"
		out += "Pipeline ID\t: " + l.id + "\n"
		if l.dryRun {
			out += "Dry Run\t\t: enabled\n"
		}
		return out
	case kindPipelineResult:
		return "\n=> " + strings.ToUpper(statusLabels[l.status]) + " (" + formatDuration(l.duration) + ")\n"
	case kindHeading:
		heading := strings.ToTitle(l.msg)
		return "\n\n" + heading + "\n" + strings.Repeat("=", lipgloss.Width(heading)) + "\n\n"
	case kindSection:
		header := l.category + ": " + l.msg
		return "\n" + header + "\n" + strings.Repeat("-", lipgloss.Width(header)) + "\n\n"
	case kindSubheading:
		return "\n" + l.msg + "\n" + strings.Repeat("-", lipgloss.Width(l.msg)) + "\n\n"
	}

	var prefix string
	switch {
	case l.level >= ErrorLevel:
		prefix = "ERROR: "
	case l.level >= WarnLevel:
		prefix = "WARNING: "
	case l.level >= InfoLevel:
	default:
		prefix = "DEBUG: "
	}
	return prefix + withNewline(l.msg)
}

// styled renders the line for a color terminal.
func (l line) styled() string {
	switch l.kind {
	case kindTitle:
		label := "── " + strings.ToTitle(l.msg) + " "
		rule := strings.Repeat("─", max(ruleWidth-lipgloss.Width(label), 3))
		return "\n\n" + titleStyle.Render(label+rule) + "\n\n"
	case kindPipelineTitle:
		return "\n" + l.pipelineBand() + "\n"
	case kindPipelineResult:
		statusStyle := statusStyles[l.status]
		label := "┗ " + statusSymbols[l.status] + " " + statusLabels[l.status]
		duration := " · " + formatDuration(l.duration) + " "
		rule := strings.Repeat("━", max(ruleWidth-lipgloss.Width(label)-lipgloss.Width(duration), 3))
		return "\n" + statusStyle.Render(label) + detailStyle.Render(duration) + statusStyle.Render(rule) + "\n"
	case kindHeading:
		return "\n\n" + headingStyle.Render(strings.ToTitle(l.msg)) + "\n" +
			ruleStyle.Render(strings.Repeat("═", lipgloss.Width(strings.ToTitle(l.msg)))) + "\n\n"
	case kindSection:
		width := lipgloss.Width(l.category + ": " + l.msg)
		return "\n" + categoryStyle.Render(l.category+":") + " " + nameStyle.Render(l.msg) + "\n" +
			ruleStyle.Render(strings.Repeat("─", width)) + "\n\n"
	case kindSubheading:
		return "\n" + nameStyle.Render(l.msg) + "\n" +
			ruleStyle.Render(strings.Repeat("─", lipgloss.Width(l.msg))) + "\n\n"
	}

	var prefix string
	switch {
	case l.level >= ErrorLevel:
		prefix = errorStyle.Render("ERROR") + ": "
	case l.level >= WarnLevel:
		prefix = warningStyle.Render("WARNING") + ": "
	case l.level >= InfoLevel:
	default:
		prefix = debugStyle.Render("DEBUG") + ": "
	}
	return prefix + withNewline(l.msg)
}

var (
	statusSymbols = map[Status]string{
		StatusUnchanged: "✔",
		StatusChanged:   "⚠",
		StatusFailed:    "✗",
		StatusSkipped:   "-",
	}
	statusLabels = map[Status]string{
		StatusUnchanged: "nothing changed",
		StatusChanged:   "changed",
		StatusFailed:    "failed",
		StatusSkipped:   "skipped",
	}
)

// counter returns the position of the pipeline within the run, such as "3/12".
func (l line) counter() string {
	if l.total == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", l.position, l.total)
}

// pipelineBand renders the opening of a pipeline as a block of colored rows: no other line has
// a background, so pipelines stand out when scrolling. The block grows rather than cutting the
// pipeline ID, which must stay complete to be used with --pipeline-ids.
func (l line) pipelineBand() string {
	title := "  " + l.msg
	if counter := l.counter(); counter != "" {
		title = "  " + counter + "  " + l.msg
	}

	mode := ""
	if l.dryRun {
		mode = "dry run  "
	}

	id := ""
	if l.id != "" {
		id = "  Pipeline ID: " + l.id + "  "
		// On a narrow terminal, the label goes first so the ID still fits
		if lipgloss.Width(id) > ruleWidth {
			id = "  " + l.id + "  "
		}
	}

	width := max(ruleWidth, lipgloss.Width(title)+2+lipgloss.Width(mode), lipgloss.Width(id))
	pad := func(text string) string {
		return text + strings.Repeat(" ", width-lipgloss.Width(text))
	}

	rows := []string{
		bandStyle.Render(pad("")),
		bandStyle.Render(pad(title + strings.Repeat(" ", width-lipgloss.Width(title)-lipgloss.Width(mode)) + mode)),
	}
	if id != "" {
		rows = append(rows, bandIDStyle.Render(pad(id)))
	}
	rows = append(rows, bandStyle.Render(pad("")))

	return strings.Join(rows, "\n") + "\n"
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func withNewline(msg string) string {
	if strings.HasSuffix(msg, "\n") {
		return msg
	}
	return msg + "\n"
}
