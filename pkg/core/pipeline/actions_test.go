package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

type mockActionHandler struct {
	cleanCounter int
}

func (m *mockActionHandler) CreateAction(ctx context.Context, report *reports.Action, resetDescription bool) error {
	return nil
}

func (m *mockActionHandler) CleanAction(ctx context.Context, report *reports.Action) error {
	m.cleanCounter++
	return nil
}

func (m *mockActionHandler) CheckActionExist(ctx context.Context, report *reports.Action) error {
	return nil
}

func TestRunCleanActions(t *testing.T) {
	testdata := []struct {
		name string
		// published reports if the action has been created or updated by the current execution
		published bool
		dryRun    bool
		// expectedCleanCounter reports how many times the action cleanup ran
		expectedCleanCounter int
	}{
		{
			name:                 "action opened by a previous execution must be cleaned",
			expectedCleanCounter: 1,
		},
		{
			/*
				An action published by the current execution must not be cleaned by that
				same execution as the cleanup relies on remote information which may not
				reflect yet what Updatecli just published.
			*/
			name:                 "action published by the current execution must not be cleaned",
			published:            true,
			expectedCleanCounter: 0,
		},
		{
			name:                 "nothing is cleaned while running in dry run",
			dryRun:               true,
			expectedCleanCounter: 0,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			handler := mockActionHandler{}

			p := Pipeline{
				Name: "test",
				Targets: map[string]target.Target{
					"default": {},
				},
				Actions: map[string]action.Action{
					"default": {
						Handler:   &handler,
						Published: tt.published,
					},
				},
			}
			p.Options.Target.DryRun = tt.dryRun

			gotErr := p.RunCleanActions(context.Background())

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedCleanCounter, handler.cleanCounter)
		})
	}
}
