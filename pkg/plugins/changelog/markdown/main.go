package markdown

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	html2md "github.com/JohannesKaufmann/html-to-markdown"
)

type Section struct {
	title           string
	astNodes        []ast.Node
	descriptionHTML string
	descriptionMD   string
}

type Sections []Section

func ParseMarkdown(data []byte) (Sections, error) {

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	doc := md.Parser().Parse(text.NewReader(data))

	sections := Sections{}

	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {

		if !entering || node.Type() == ast.TypeDocument {
			return ast.WalkContinue, nil
		}

		if _, ok := node.(*ast.Heading); ok && node.(*ast.Heading).Level == 2 {
			sections = append(
				sections,
				Section{
					title:    string(node.Lines().Value(data)),
					astNodes: nil,
				},
			)
		}
		if len(sections) > 0 {
			segment := sections[len(sections)-1].astNodes
			segment = append(segment, node)
			sections[len(sections)-1].astNodes = segment
		}
		return ast.WalkSkipChildren, nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking the markdown: %s", err)
	}

	for i := range sections {
		doc := ast.NewDocument()
		for _, node := range sections[i].astNodes {
			if node.Kind() == ast.KindHeading {
				switch node.(*ast.Heading).Level {
				case 1, 2:
					// Release description shouldn't including release title
				default:
					doc.AppendChild(doc, node)
				}
			} else {
				doc.AppendChild(doc, node)
			}

		}
		var h strings.Builder

		if err := md.Renderer().Render(&h, data, doc); err != nil {
			return nil, fmt.Errorf("rendering the markdown: %s", err)
		}

		converter := html2md.NewConverter("", true, nil)
		mdSection, err := converter.ConvertString(h.String())
		if err != nil {
			return nil, fmt.Errorf("converting the html to markdown: %s", err)
		}

		sections[i].descriptionHTML = h.String()
		sections[i].descriptionMD = mdSection
	}

	return sections, nil
}

// GetSectionAsMarkdown returns the section with the given title.
func (s Sections) GetSectionAsMarkdown(title string) string {
	titles := []string{}

	for i := range s {
		titles = append(titles, s[i].title)
		if s[i].title == title {
			return s[i].descriptionMD
		}
	}

	logrus.Debugf("Version %q not found from Changelog versions %v", title, titles)
	return ""
}

// GetSectionAsHTML returns the section with the given title.
func (s Sections) GetSectionAsHTML(title string) string {
	titles := []string{}

	for i := range s {
		titles = append(titles, s[i].title)
		if s[i].title == title {
			return s[i].descriptionHTML
		}
	}

	logrus.Debugf("Version %q not found from Changelog versions %v", title, titles)
	return ""
}
