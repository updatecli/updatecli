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
			expectedOutput: `<Action id="1234">
    <h2>Test Title</h2>
    <details id="4567">
        <h3>Target One</h3>
        <details>
            <summary>1.0.0</summary>
        </details>
        <details>
            <summary>1.0.1</summary>
        </details>
    </details>
    <details id="4567">
        <h3>Target Two</h3>
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
		expectedOutput Action
	}{
		{
			name: "Default working situation",
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
			report: `<action id="1234">
    <h2>Test Title</h2>
    <p></p>
    <details id="4567">
        <h3>Target One</h3>
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
        <h3>Target Two</h3>
        <p></p>
    </details>
</action>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOutput Action
			err := Unmarshal([]byte(tt.report), &gotOutput)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedOutput, gotOutput)
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
			err := tt.report.Sort()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, tt.report)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.report1.Merge(&tt.report2)
			err := tt.report1.Sort()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, tt.report1)
		})
	}
}
