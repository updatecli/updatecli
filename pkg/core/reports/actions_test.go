package reports

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionsMerge(t *testing.T) {
	tests := []struct {
		name                string
		oldReport           Actions
		newReport           Actions
		expectedFinalReport Actions
	}{
		{
			name: "Default none situation",
			oldReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
		},
		{
			name: "Test target merge",
			oldReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
		},
		{
			name: "Test Pipeline merge",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Old Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1235",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Old Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
		},
		{
			name: "Test Pipeline merge scenario 2",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1235",
						PipelineTitle: "Old Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Old Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Test Pipeline merge scenario 3",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Old Pipeline 1",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1236",
						PipelineTitle: "Old Pipeline 2",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1235",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Old Pipeline 1",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
					{
						ID:            "1236",
						PipelineTitle: "Old Pipeline 2",
						Targets: []ActionTarget{
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
		},
		{
			name:      "No merge needed",
			oldReport: Actions{},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1235",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1235",
						PipelineTitle: "New Pipeline",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
								Changelogs: []ActionTargetChangelog{
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Update target title numbers match",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
		},
		{
			name: "Update pipeline title with new report having more targets",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
		},
		{
			name: "Update pipeline title with old report having more targets",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4569",
								Title: "New Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "New Target Three",
							},
						},
					},
				},
			},
		},
		{
			name: "Update target title numbers match and old report having more pipelines",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with new report having more targets and old report having more pipelines",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with old report having more targets and old report having more pipelines",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4569",
								Title: "New Target Three",
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "New Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Update target title numbers match and new report having more pipelines",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with new report having more targets and new report having more pipelines",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "New Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with old report having more targets and new report having more pipelines",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "Target Three",
							},
						},
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4569",
								Title: "New Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
						Targets: []ActionTarget{
							{
								ID:    "4567",
								Title: "Target One",
							},
							{
								ID:    "4568",
								Title: "Target Two",
							},
							{
								ID:    "4569",
								Title: "New Target Three",
							},
						},
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with old report having same number actions",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Title",
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with old report having more actions",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
					},
					{
						ID:            "1235",
						PipelineTitle: "Old Title",
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Title",
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Title",
					},
					{
						ID:            "1235",
						PipelineTitle: "Old Title",
					},
				},
			},
		},
		{
			name: "Update pipeline title with new report having more actions",
			oldReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "Test Title",
					},
				},
			},
			newReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Title",
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
			expectedFinalReport: Actions{
				Actions: []Action{
					{
						ID:            "1234",
						PipelineTitle: "New Title",
					},
					{
						ID:            "1235",
						PipelineTitle: "Other Title",
					},
				},
			},
		},
		{
			name: "Test changelog merge with old report having more items",
			oldReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			newReport: Actions{
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
										Title: "1.0.2",
									},
								},
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
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
										Title: "1.0.2",
									},
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "Test changelog merge with new report having more items",
			oldReport: Actions{
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
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
			newReport: Actions{
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
										Title: "1.0.2",
									},
									{
										Title: "1.0.3",
									},
									{
										Title: "1.0.4",
									},
								},
							},
						},
					},
				},
			},
			expectedFinalReport: Actions{
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
										Title: "1.0.4",
									},
									{
										Title: "1.0.3",
									},
									{
										Title: "1.0.2",
									},
									{
										Title: "1.0.1",
									},
									{
										Title: "1.0.0",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.newReport.Merge(&tt.oldReport)
			tt.newReport.sort()
			assert.Equal(t, tt.expectedFinalReport, tt.newReport)
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
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
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
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
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
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
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
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
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
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
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
			name: "No merge needed",
			newReport: `<Actions>
    <action id="1235">
        <h3>New Pipeline</h3>
        <details id="4567">
            <summary>Target One</summary>
            <details>
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
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
                <summary>1.0.1</summary>
            </details>
            <details>
                <summary>1.0.0</summary>
            </details>
        </details>
    </action>
</Actions>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFinalReport := MergeFromString(tt.oldReport, tt.newReport)
			assert.Equal(t, tt.expectedFinalReport, gotFinalReport)
		})
	}
}

func TestMergeFromMarkdown(t *testing.T) {
	tests := []struct {
		name                string
		oldReport           string
		newReport           string
		expectedFinalReport string
	}{
		{
			name: "Default none situation",
			oldReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
			newReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
			expectedFinalReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
		},
		{
			name: "Test target merge",
			oldReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0`,
			newReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
			expectedFinalReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
		},
		{
			name: "Test that old report includes unexpected text",
			oldReport: `This is not a markdown expected format

# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
			newReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
			expectedFinalReport: `# Test Title

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
		},
		{
			name: "Test Pipeline merge",
			oldReport: `# Old Pipeline

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0`,
			newReport: `# New Pipeline

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
			expectedFinalReport: `# New Pipeline

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0

## Target Two

Target ID: ` + "`" + `4568` + "`" + `

## Target Three

Target ID: ` + "`" + `4569` + "`",
		},
		{
			name: "No merge needed",
			newReport: `# New Pipeline

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0`,
			oldReport: "",
			expectedFinalReport: `# New Pipeline

Pipeline ID: ` + "`" + `1234` + "`" + `

## Target One

Target ID: ` + "`" + `4567` + "`" + `

### 1.0.1

### 1.0.0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFinalReport, err := MergeFromMarkdown(tt.oldReport, tt.newReport)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedFinalReport, gotFinalReport)
		})
	}
}
