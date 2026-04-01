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
