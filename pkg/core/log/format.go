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
	// id, dryRun, status and duration describe a pipeline
	id       string
	dryRun   bool
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
	pipelineStyle lipgloss.Style
	barStyle      lipgloss.Style
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
	pipelineStyle = r.NewStyle().Bold(true)
	barStyle = r.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
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
		border := strings.Repeat("#", len(l.msg)+4)
		out := "\n" + border + "\n# " + strings.ToTitle(l.msg) + " #\n" + border + "\n\n"
		out += "Pipeline ID\t: " + l.id + "\n"
		if l.dryRun {
			out += "Dry Run\t\t: enabled\n"
		}
		return out
	case kindPipelineResult:
		return "\n=> " + strings.ToUpper(statusLabels[l.status]) + " (" + formatDuration(l.duration) + ")\n"
	case kindHeading:
		return "\n\n" + strings.ToTitle(l.msg) + "\n" + strings.Repeat("=", len(l.msg)) + "\n\n"
	case kindSection:
		return "\n" + l.category + ": " + l.msg + "\n" + strings.Repeat("-", len(l.msg)+len(l.category)+2) + "\n\n"
	case kindSubheading:
		return "\n" + l.msg + "\n" + strings.Repeat("-", len(l.msg)) + "\n\n"
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
		rule := strings.Repeat("─", max(ruleWidth-len([]rune(label)), 3))
		return "\n\n" + titleStyle.Render(label+rule) + "\n\n"
	case kindPipelineTitle:
		bar := barStyle.Render("┃") + " "
		out := "\n" + bar + pipelineStyle.Render(l.msg) + "\n"
		if details := l.pipelineDetails(); details != "" {
			out += bar + detailStyle.Render(details) + "\n"
		}
		return out + "\n"
	case kindPipelineResult:
		statusStyle := statusStyles[l.status]
		return "\n" + statusStyle.Render("┗ "+statusSymbols[l.status]+" "+statusLabels[l.status]) +
			detailStyle.Render(" · "+formatDuration(l.duration)) + "\n"
	case kindHeading:
		return "\n\n" + headingStyle.Render(strings.ToTitle(l.msg)) + "\n" +
			ruleStyle.Render(strings.Repeat("═", len([]rune(l.msg)))) + "\n\n"
	case kindSection:
		width := len([]rune(l.msg)) + len(l.category) + 2
		return "\n" + categoryStyle.Render(l.category+":") + " " + nameStyle.Render(l.msg) + "\n" +
			ruleStyle.Render(strings.Repeat("─", width)) + "\n\n"
	case kindSubheading:
		return "\n" + nameStyle.Render(l.msg) + "\n" +
			ruleStyle.Render(strings.Repeat("─", len([]rune(l.msg)))) + "\n\n"
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

// pipelineDetails returns the short pipeline ID and the dry run mode, the full ID is in the report.
func (l line) pipelineDetails() string {
	var details []string
	if l.id != "" {
		details = append(details, l.id[:min(len(l.id), 8)])
	}
	if l.dryRun {
		details = append(details, "dry run")
	}
	return strings.Join(details, " · ")
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
