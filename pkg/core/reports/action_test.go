package reports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLReportsString(t *testing.T) {
	tests := []struct {
		name           string
		report         Action
		expectedOutput string
	}{
		{
			name: "Default working situation",
			report: Action{
				ID:            "1234",
				Title:         "Action Title",
				PipelineTitle: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:    "4567",
						Title: "Target Two",
					},
				},
			},
			expectedOutput: `<Action id="1234">
    <h3>Test Title</h3>
    <details id="4567">
        <summary>Target One</summary>
        <details>
            <summary>1.0.0</summary>
        </details>
        <details>
            <summary>1.0.1</summary>
        </details>
    </details>
    <details id="4567">
        <summary>Target Two</summary>
    </details>
</Action>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedOutput, tt.report.String())
		})
	}
}

func TestHTMLUnmarshal(t *testing.T) {
	tests := []struct {
		name           string
		report         string
		expectedOutput Actions
	}{
		{
			name: "Default working situation",
			expectedOutput: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title:       "1.0.0",
										Description: "",
									},
									{
										Title:       "1.0.1",
										Description: "",
									},
								},
							},
							{
								ID:    "4567",
								Title: "Target Two",
							},
						},
					},
				},
			},
			report: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
            <details>
                <summary>1.0.0</summary>
            </details>
            <details>
                <summary>1.0.1</summary>
            </details>
        </details>
        <details id="4567">
            <summary>Target Two</summary>
        </details>
    </action>
</Actions>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOutput Actions
			err := unmarshal([]byte(tt.report), &gotOutput)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedOutput, gotOutput)

			// test round trip
			assert.Equal(t, tt.report, gotOutput.Actions[0].ToActionsString())
		})
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		name           string
		report         Action
		expectedOutput Action
	}{
		{
			name: "Canonical scenario, both are matching",
			expectedOutput: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:    "4567",
						Title: "Target Two",
					},
				},
			},
			report: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:    "4567",
						Title: "Target Two",
					},
				},
			},
		},
		{
			name: "Should must be reorder",
			expectedOutput: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:    "4567",
						Title: "Target Two",
					},
				},
			},
			report: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.1",
								Description: "",
							},
							{
								Title:       "1.0.0",
								Description: "",
							},
						},
					},
					{
						ID:    "4567",
						Title: "Target Two",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.report.sort()
			assert.Equal(t, tt.expectedOutput, tt.report)
		})
	}
}

func TestActionMerge(t *testing.T) {
	tests := []struct {
		name           string
		report1        Action
		report2        Action
		expectedOutput Action
	}{
		{
			name: "Should must be merged",
			report1: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
						},
					},
					{
						ID:    "4568",
						Title: "Target Two",
					},
				},
			},
			report2: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:    "4568",
						Title: "Target Two",
					},
				},
			},
			expectedOutput: Action{
				ID:    "1234",
				Title: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:    "4568",
						Title: "Target Two",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.report1.Merge(&tt.report2, true)
			tt.report1.sort()
			assert.Equal(t, tt.expectedOutput, tt.report1)
		})
	}
}

func TestToActionsMarkdownString(t *testing.T) {
	tests := []struct {
		name           string
		report         Action
		expectedOutput string
	}{
		{
			name: "Default working situation",
			report: Action{
				ID:            "1234",
				Title:         "Action Title",
				PipelineTitle: "Test Title",
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:          "4567",
						Title:       "Target Two",
						Description: "Description",
					},
					{
						ID:    "4567",
						Title: "Target Three",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "Description",
							},
						},
					},
				},
			},
			expectedOutput: `# Test Title

## Target One

### 1.0.0

### 1.0.1

## Target Two

Description

## Target Three

### 1.0.0

` + "```" + `
Description
` + "```",
		},
		{
			name: "Multiline",
			report: Action{
				ID:            "1234",
				Title:         "Action Title",
				PipelineTitle: "Test Title",
				Targets: []ActionTarget{
					{
						ID:          "4567",
						Title:       "Target One",
						Description: "Something happened\n\t* to this file",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "# v1.0.0\n\nfeat: something cool",
							},
							{
								Title:       "1.0.1",
								Description: "# v1.0.1\n\nfix: something fixed",
							},
						},
					},
					{
						ID:          "4568",
						Title:       "Target Two",
						Description: "Something happened\n\t* to other this file",
					},
					{
						ID:    "4569",
						Title: "Target Three",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "# v1.0.0\n\nfeat: something cool",
							},
						},
					},
				},
			},
			expectedOutput: `# Test Title

## Target One

Something happened

* to this file

### 1.0.0

` + "```" + `
# v1.0.0

feat: something cool
` + "```" + `

### 1.0.1

` + "```" + `
# v1.0.1

fix: something fixed
` + "```" + `

## Target Two

Something happened

* to other this file

## Target Three

### 1.0.0

` + "```" + `
# v1.0.0

feat: something cool
` + "```",
		},
		{
			name: "with PipelineUrl",
			report: Action{
				ID:            "1234",
				Title:         "Action Title",
				PipelineTitle: "Test Title",
				PipelineURL: &PipelineURL{
					URL:  "https://www.updatecli.io/",
					Name: "updatecli",
				},
				Targets: []ActionTarget{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "",
							},
							{
								Title:       "1.0.1",
								Description: "",
							},
						},
					},
					{
						ID:          "4567",
						Title:       "Target Two",
						Description: "Description",
					},
					{
						ID:    "4567",
						Title: "Target Three",
						Changelogs: []ActionTargetChangelog{
							{
								Title:       "1.0.0",
								Description: "Description",
							},
						},
					},
				},
			},
			expectedOutput: `# Test Title

## Target One

### 1.0.0

### 1.0.1

## Target Two

Description

## Target Three

### 1.0.0

` + "```" + `
Description
` + "```" + `

[updatecli](https://www.updatecli.io/)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedOutput, tt.report.ToActionsMarkdownString())
		})
	}
}
