package pyproject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePEP508(t *testing.T) {
	testdata := []struct {
		name      string
		input     string
		expected  pythonDependency
		expectErr bool
	}{
		{
			name:     "Simple with >=",
			input:    "requests>=2.28",
			expected: pythonDependency{Name: "requests", Constraint: ">=2.28", Version: "2.28"},
		},
		{
			name:     "Multiple constraints",
			input:    "flask>=3.0,<4",
			expected: pythonDependency{Name: "flask", Constraint: ">=3.0,<4", Version: "3.0"},
		},
		{
			name:     "Compatible release",
			input:    "pydantic~=2.0",
			expected: pythonDependency{Name: "pydantic", Constraint: "~=2.0", Version: "2.0"},
		},
		{
			name:     "Exact pin",
			input:    "black==24.1.0",
			expected: pythonDependency{Name: "black", Constraint: "==24.1.0", Version: "24.1.0"},
		},
		{
			name:     "Constraint between parentheses",
			input:    "requests (==2.25.0)",
			expected: pythonDependency{Name: "requests", Constraint: "==2.25.0", Version: "2.25.0"},
		},
		{
			name:     "Extras and constraints between parentheses",
			input:    "requests[socks] ( >=2.0, <3 ) ; python_version >= '3.12'",
			expected: pythonDependency{Name: "requests", Extras: "socks", Constraint: ">=2.0, <3", Version: "2.0"},
		},
		{
			name:     "No version",
			input:    "numpy",
			expected: pythonDependency{Name: "numpy"},
		},
		{
			name:     "Extras",
			input:    "black[jupyter]>=24.0",
			expected: pythonDependency{Name: "black", Extras: "jupyter", Constraint: ">=24.0", Version: "24.0"},
		},
		{
			name:     "Env markers stripped",
			input:    "pywin32>=300; sys_platform == 'win32'",
			expected: pythonDependency{Name: "pywin32", Constraint: ">=300", Version: "300"},
		},
		{
			name:      "URL dep returns error",
			input:     "package @ https://example.com/pkg.tar.gz",
			expectErr: true,
		},
		{
			name:      "URL dep without spaces around @",
			input:     "package@https://example.com/pkg.tar.gz",
			expectErr: true,
		},
		{
			name:      "URL dep with space only before @",
			input:     "package @https://example.com/pkg.tar.gz",
			expectErr: true,
		},
		{
			name:      "URL dep with space only after @",
			input:     "package@ https://example.com/pkg.tar.gz",
			expectErr: true,
		},
		{
			name:     "Hyphenated name",
			input:    "my-package>=1.0",
			expected: pythonDependency{Name: "my-package", Constraint: ">=1.0", Version: "1.0"},
		},
		{
			name:     "Dotted name",
			input:    "zope.interface>=5.0",
			expected: pythonDependency{Name: "zope.interface", Constraint: ">=5.0", Version: "5.0"},
		},
		{
			name:     "PEP 440 beta pre-release",
			input:    "otel-instrumentation>=0.51b0",
			expected: pythonDependency{Name: "otel-instrumentation", Constraint: ">=0.51b0", Version: "0.51b0"},
		},
		{
			name:     "PEP 440 alpha pre-release",
			input:    "pkg>=1.0a1",
			expected: pythonDependency{Name: "pkg", Constraint: ">=1.0a1", Version: "1.0a1"},
		},
		{
			name:     "PEP 440 rc pre-release",
			input:    "pkg>=2.0rc1",
			expected: pythonDependency{Name: "pkg", Constraint: ">=2.0rc1", Version: "2.0rc1"},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePEP508(tt.input)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Name, got.Name)
			assert.Equal(t, tt.expected.Extras, got.Extras)
			assert.Equal(t, tt.expected.Constraint, got.Constraint)
			assert.Equal(t, tt.expected.Version, got.Version)
		})
	}
}

func TestDependencyGroupRequirements(t *testing.T) {
	testdata := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "Only strings",
			input:    []interface{}{"pytest>=8.0", "ruff"},
			expected: []string{"pytest>=8.0", "ruff"},
		},
		{
			name:     "include-group entries are skipped",
			input:    []interface{}{map[string]interface{}{"include-group": "test"}, "ruff>=0.4"},
			expected: []string{"ruff>=0.4"},
		},
		{
			name:     "Empty group",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Array of tables only holding include-group entries",
			input:    []map[string]interface{}{{"include-group": "test"}, {"include-group": "lint"}},
			expected: nil,
		},
		{
			name:     "Group that isn't an array is skipped",
			input:    map[string]interface{}{"pytest": ">=8.0"},
			expected: nil,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, dependencyGroupRequirements(tt.input))
		})
	}
}

func TestDeclaredLowerBound(t *testing.T) {
	testdata := []struct {
		constraint string
		expected   string
	}{
		{constraint: ">=2.28", expected: "2.28"},
		{constraint: ">=2.28,<3", expected: "2.28"},
		{constraint: "<3, >= 2.28", expected: "2.28"},
		{constraint: "~=1.4.2", expected: "1.4.2"},
		{constraint: "==1.2.3", expected: "1.2.3"},
		{constraint: "==1.*", expected: "1"},
		{constraint: ">=2.28.post1", expected: "2.28"},
		{constraint: ">=1.0.dev0", expected: "1.0"},
		{constraint: ">=2.0rc1,<3", expected: "2.0"},
		{constraint: ">1.0", expected: "1.0"},
		// Constraints combined from several declarations start from the highest lower bound
		{constraint: ">=2.0,<2.30,>=2.10", expected: "2.10"},
		{constraint: "<3.0", expected: ""},
		{constraint: "!=1.5", expected: ""},
		{constraint: "", expected: ""},
		{constraint: "<=2.0,!=1.5", expected: ""},
	}

	for _, tt := range testdata {
		t.Run(tt.constraint, func(t *testing.T) {
			assert.Equal(t, tt.expected, declaredLowerBound(tt.constraint))
		})
	}
}
