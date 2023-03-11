package reports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLReportsString(t *testing.T) {

	tests := []struct {
		name           string
		report         htmlReport
		expectedOutput string
	}{
		{
			name: "Default working situation",
			report: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			expectedOutput: `<htmlReport id="1234">
    <h2>Test Title</h2>
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
</htmlReport>`,
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
		expectedOutput htmlReport
	}{
		{
			name: "Default working situation",
			expectedOutput: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			report: `<htmlReport id="1234">
    <h2>Test Title</h2>
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
</htmlReport>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOutput htmlReport
			err := Unmarshal([]byte(tt.report), &gotOutput)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedOutput, gotOutput)
		})
	}
}

func TestSort(t *testing.T) {

	tests := []struct {
		name           string
		report         htmlReport
		expectedOutput htmlReport
	}{
		{
			name: "Canonical scenario, both are matching",
			expectedOutput: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			report: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			expectedOutput: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			report: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
		report1        htmlReport
		report2        htmlReport
		expectedOutput htmlReport
	}{
		{
			name: "Should must be merged",
			report1: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			report2: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
			expectedOutput: htmlReport{
				ID:    "1234",
				Title: "Test Title",
				Targets: []targetHTMLReport{
					{
						ID:    "4567",
						Title: "Target One",
						Changelogs: []HTMLChangelog{
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
