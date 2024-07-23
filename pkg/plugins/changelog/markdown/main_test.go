package markdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type releaseTest struct {
	title           string
	descriptionHTML string
}

type releaseTestData struct {
	name        string
	data        string
	expectedXML []releaseTest
	wantErr     bool
}

var (
	tests = []releaseTestData{
		{
			name: "Loading markdown syntax",
			data: `
# Changelog

## 0.1.0

	xxx
yyy

[xe](https://localhost)

* xxxx
* yyyy

## 0.2.0

eod

`,
			expectedXML: []releaseTest{
				{
					title: "0.1.0",
					descriptionHTML: `<pre><code>xxx
</code></pre>
<p>yyy</p>
<p><a href="https://localhost">xe</a></p>
<ul>
<li>xxxx</li>
<li>yyyy</li>
</ul>
`,
				},
				{
					title: "0.2.0",
					descriptionHTML: `<p>eod</p>
`,
				},
			},
		},
		{
			name: "Wrongly parsing asciidoctor",
			data: `
= Changelog

== 0.1.0

* Add tests

== 0.2.0

* Initial release
`,
			expectedXML: []releaseTest{},
		},
		{
			name: "Loading old markdown syntax",
			data: `
Rhai Release Notes
==================

Version 1.20.0
--------------

* (Fuzzing) An integer-overflow bug from an inclusive range in the bits iterator is fixed.
`,
			expectedXML: []releaseTest{
				{
					title: "Version 1.20.0",
					descriptionHTML: `<ul>
<li>(Fuzzing) An integer-overflow bug from an inclusive range in the bits iterator is fixed.</li>
</ul>
`,
				},
			},
		},
	}
)

func TestParseMarkdown(t *testing.T) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData, err := ParseMarkdown([]byte(tt.data))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			for i := range xmlData {
				require.Equal(t, tt.expectedXML[i].title, xmlData[i].title)
				require.Equal(t, tt.expectedXML[i].descriptionHTML, xmlData[i].descriptionHTML)
			}
		})
	}
}

func TestGetSection(t *testing.T) {

	data := `
# Changelog

## 0.2.0

* Introduce new tests

## 0.1.0

* Initial release
`

	sectionData := []struct {
		name                        string
		searchTitle                 string
		expectedMarkdownDescription string
		expectedHTMLDescription     string
		expectedNil                 bool
	}{
		{
			searchTitle:                 "0.1.0",
			expectedMarkdownDescription: "- Initial release",
			expectedHTMLDescription:     "<ul>\n<li>Initial release</li>\n</ul>\n",
		},
		{
			searchTitle: "0.0.0",
			expectedNil: true,
		},
	}

	for _, tt := range sectionData {
		t.Run(tt.name, func(t *testing.T) {
			xmlData, err := ParseMarkdown([]byte(data))
			require.NoError(t, err)

			gotMarkdownSection := xmlData.GetSectionAsMarkdown(tt.searchTitle)
			gotHTMLSection := xmlData.GetSectionAsHTML(tt.searchTitle)

			if tt.expectedNil {
				require.Empty(t, gotMarkdownSection)
			} else {
				require.Equal(t, tt.expectedMarkdownDescription, gotMarkdownSection)
				require.Equal(t, tt.expectedHTMLDescription, gotHTMLSection)
			}
		})
	}
}
