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
			report: `<actions>
	<action id="1234">
	    <h3>Test Title</h3>
	    <p></p>
	    <details id="4567">
	        <summary>Target One</summary>
	        <p></p>
	        <details>
	            <summary>1.0.0</summary>
	            <p></p>
	        </details>
	        <details>
	            <summary>1.0.1</summary>
	            <p></p>
	        </details>
	    </details>
	    <details id="4567">
	        <summary>Target Two</summary>
	        <p></p>
	    </details>
	</action>
</actions>`,
		},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			var gotOutput Actions
			err := unmarshal([]byte(tests[i].report), &gotOutput)
			require.NoError(t, err)

			assert.Equal(t, tests[i].expectedOutput, gotOutput)
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

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].report.sort()
			assert.Equal(t, tests[i].expectedOutput, tests[i].report)
		})
	}
}

func TestMerge(t *testing.T) {
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

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].report1.Merge(&tests[i].report2)
			tests[i].report1.sort()
			assert.Equal(t, tests[i].expectedOutput, tests[i].report1)
		})
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		name                string
		oldReport           string
		newReport           string
		expectedFinalReport string
	}{
		{
			name: "Default none situation",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <p></p>
        <details id="4567">
            <summary>Target One</summary>
            <p></p>
            <details>
                <summary>1.0.0</summary>
                <p></p>
            </details>
            <details>
                <summary>1.0.1</summary>
                <p></p>
            </details>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
            <p></p>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
            <p></p>
        </details>
    </action>
</Actions>`,
			newReport: `<Actions>
	<action id="1234">
		<h3>Test Title</h3>
		<p></p>
		<details id="4567">
		    <summary>Target One</summary>
		    <p></p>
		    <details>
		        <summary>1.0.0</summary>
		        <p></p>
		    </details>
		    <details>
		        <summary>1.0.1</summary>
		        <p></p>
		    </details>
		</details>
		<details id="4568">
		    <summary>Target Two</summary>
		    <p></p>
		</details>
		<details id="4569">
		    <summary>Target Three</summary>
		    <p></p>
		</details>
	</action>
</Actions>`,
			expectedFinalReport: `<Actions>
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
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "Test target merge",
			oldReport: `<Actions>
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
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
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
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "Test that old report is not fully html formatted",
			oldReport: `
This is not a html formatted report
<Action id="1234">
    <h3>Test Title</h3>
    <details id="4568">
        <summary>Target Two</summary>
    </details>
    <details id="4569">
        <summary>Target Three</summary>
    </details>
</Action>`,
			newReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4568">
		    <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "Test Pipeline merge",
			oldReport: `<actions>
    <action id="1234">
        <h3>Old Pipeline</h3>
        <details id="4567">
            <summary>Target One</summary>
            <details>
                <summary>1.0.0</summary>
            </details>
            <details>
                <summary>1.0.1</summary>
            </details>
        </details>
    </action>
</actions>`,
			newReport: `<actions>
    <action id="1235">
        <h3>New Pipeline</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Old Pipeline</h3>
        <details id="4567">
            <summary>Target One</summary>
            <details>
                <summary>1.0.0</summary>
            </details>
            <details>
                <summary>1.0.1</summary>
            </details>
        </details>
    </action>
    <action id="1235">
        <h3>New Pipeline</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "Test Pipeline merge scenario 2",
			newReport: `<actions>
    <action id="1235">
    	<h3>Old Pipeline</h3>
    	<details id="4567">
    	    <summary>Target One</summary>
    	    <details>
    	        <summary>1.0.0</summary>
    	    </details>
    	    <details>
    	        <summary>1.0.1</summary>
    	    </details>
    	</details>
    </action>
</actions>`,
			oldReport: `<actions>
    <action id="1234">
        <h3>New Pipeline</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>New Pipeline</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Old Pipeline</h3>
        <details id="4567">
            <summary>Target One</summary>
            <details>
                <summary>1.0.0</summary>
            </details>
            <details>
                <summary>1.0.1</summary>
            </details>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "Test Pipeline merge scenario 3",
			newReport: `<Actions>
	<action id="1235">
		<h3>New Pipeline</h3>
		<details id="4567">
		    <summary>Target One</summary>
		    <details>
		        <summary>1.0.0</summary>
		    </details>
		    <details>
		        <summary>1.0.1</summary>
		    </details>
		</details>
	</action>
</Actions>`,
			oldReport: `<actions>
	<action id="1234">
	    <h3>Old Pipeline 1</h3>
	    <details id="4568">
	        <summary>Target Two</summary>
	    </details>
	    <details id="4569">
	        <summary>Target Three</summary>
	    </details>
	</action>
	<action id="1236">
	    <h3>Old Pipeline 2</h3>
	    <details id="4568">
	        <summary>Target Two</summary>
	    </details>
	    <details id="4569">
	        <summary>Target Three</summary>
	    </details>
	</action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Old Pipeline 1</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>New Pipeline</h3>
        <details id="4567">
            <summary>Target One</summary>
            <details>
                <summary>1.0.0</summary>
            </details>
            <details>
                <summary>1.0.1</summary>
            </details>
        </details>
    </action>
    <action id="1236">
        <h3>Old Pipeline 2</h3>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "No merge needed",
			newReport: `<Actions>
	<action id="1235">
		<h3>New Pipeline</h3>
		<details id="4567">
		    <summary>Target One</summary>
		    <details>
		        <summary>1.0.0</summary>
		    </details>
		    <details>
		        <summary>1.0.1</summary>
		    </details>
		</details>
	</action>
</Actions>`,
			oldReport: "",
			expectedFinalReport: `<Actions>
	<action id="1235">
		<h3>New Pipeline</h3>
		<details id="4567">
		    <summary>Target One</summary>
		    <details>
		        <summary>1.0.0</summary>
		    </details>
		    <details>
		        <summary>1.0.1</summary>
		    </details>
		</details>
	</action>
</Actions>`,
		},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			gotFinalReport := MergeFromString(tests[i].oldReport, tests[i].newReport)
			assert.Equal(t, tests[i].expectedFinalReport, gotFinalReport)
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
