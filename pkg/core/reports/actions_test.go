package reports

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{
			name: "Update target title numbers match",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
			newReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <p></p>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
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
        <details id="4567">
            <summary>New Target One</summary>
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
			name: "Update pipeline title with new report having more targets",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
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
            <summary>New Target One</summary>
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
			name: "Update pipeline title with old report having more targets",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4569">
            <summary>New Target Three</summary>
        </details>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>New Target Three</summary>
        </details>
    </action>
</Actions>`,
		},
		{
			name: "Update target title numbers match and old report having more pipelines",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
			newReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <p></p>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
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
        <details id="4567">
            <summary>New Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with new report having more targets and old report having more pipelines",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
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
            <summary>New Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with old report having more targets and old report having more pipelines",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4569">
            <summary>New Target Three</summary>
        </details>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>New Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update target title numbers match and new report having more pipelines",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
			newReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <p></p>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with new report having more targets and new report having more pipelines",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>New Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with old report having more targets and new report having more pipelines",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>Target Three</summary>
        </details>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4569">
            <summary>New Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
        <details id="4567">
            <summary>Target One</summary>
        </details>
        <details id="4568">
            <summary>Target Two</summary>
        </details>
        <details id="4569">
            <summary>New Target Three</summary>
        </details>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with old report having same number actions",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>New Title</h3>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>New Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with old report having more actions",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
    </action>
    <action id="1235">
        <h3>Old Title</h3>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>New Title</h3>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>New Title</h3>
    </action>
    <action id="1235">
        <h3>Old Title</h3>
    </action>
</Actions>`,
		},
		{
			name: "Update pipeline title with new report having more actions",
			oldReport: `<Actions>
    <action id="1234">
        <h3>Test Title</h3>
    </action>
</Actions>`,
			newReport: `<actions>
    <action id="1234">
        <h3>New Title</h3>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
    </action>
</actions>`,
			expectedFinalReport: `<Actions>
    <action id="1234">
        <h3>New Title</h3>
    </action>
    <action id="1235">
        <h3>Other Title</h3>
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
