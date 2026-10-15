package log

import (
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// kind tells how a log line is rendered.
type kind int

const (
	kindMessage kind = iota
	kindTitle
	kindPipelineTitle
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
}

var (
	styleOnce sync.Once
	styled    bool

	debugStyle    lipgloss.Style
	warningStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	titleStyle    lipgloss.Style
	pipelineStyle lipgloss.Style
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
	})
	return styled
}

func initStyles(r *lipgloss.Renderer) {
	// The basic ANSI colors follow the terminal theme, light or dark
	debugStyle = r.NewStyle().Foreground(lipgloss.Color("6")).Faint(true)
	warningStyle = r.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	errorStyle = r.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	titleStyle = r.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(0, 1)
	pipelineStyle = r.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("4")).
		Padding(0, 1)
	headingStyle = r.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	categoryStyle = r.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	nameStyle = r.NewStyle().Bold(true)
	ruleStyle = r.NewStyle().Faint(true)
}

// plain renders the line as plain text, as stored in reports.
func (l line) plain() string {
	switch l.kind {
	case kindTitle:
		border := strings.Repeat("+", len(l.msg)+4)
		return "\n\n" + border + "\n+ " + strings.ToTitle(l.msg) + " +\n" + border + "\n\n"
	case kindPipelineTitle:
		border := strings.Repeat("#", len(l.msg)+4)
		return "\n" + border + "\n# " + strings.ToTitle(l.msg) + " #\n" + border + "\n\n"
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
		return "\n" + titleStyle.Render(strings.ToTitle(l.msg)) + "\n\n"
	case kindPipelineTitle:
		return "\n" + pipelineStyle.Render(l.msg) + "\n\n"
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

func withNewline(msg string) string {
	if strings.HasSuffix(msg, "\n") {
		return msg
	}
	return msg + "\n"
}
