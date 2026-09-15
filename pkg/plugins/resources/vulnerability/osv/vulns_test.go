package osv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnownVulnerabilities(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		responses     map[string]string
		statusCode    int
		expectedIDs   []string
		expectedError bool
	}{
		{
			name:        "Aliases are merged and withdrawn records dropped",
			spec:        Spec{Ecosystem: "PyPI", Name: "jinja2"},
			responses:   map[string]string{"3.1.2|": jinja2At312},
			expectedIDs: []string{"GHSA-cpwx-vrp4-4pq7", "GHSA-h75v-3vvj-5mfj", "GHSA-high-0000-0000", "PYSEC-2026-9999"},
		},
		{
			name:        "Ignore matches an alias",
			spec:        Spec{Ecosystem: "PyPI", Name: "jinja2", Ignore: []string{"cve-2025-27516", "PYSEC-2026-1474"}},
			responses:   map[string]string{"3.1.2|": jinja2At312},
			expectedIDs: []string{"GHSA-high-0000-0000", "PYSEC-2026-9999"},
		},
		{
			name:        "Minimum severity keeps vulnerabilities with an unknown severity",
			spec:        Spec{Ecosystem: "PyPI", Name: "jinja2", MinSeverity: "high"},
			responses:   map[string]string{"3.1.2|": jinja2At312},
			expectedIDs: []string{"GHSA-high-0000-0000", "PYSEC-2026-9999"},
		},
		{
			name: "Pages are followed",
			spec: Spec{Ecosystem: "PyPI", Name: "jinja2"},
			responses: map[string]string{
				"3.1.2|":      response("page2", ghsaSandbox, pysecSandbox),
				"3.1.2|page2": response("", ghsaXSS),
			},
			expectedIDs: []string{"GHSA-cpwx-vrp4-4pq7", "GHSA-h75v-3vvj-5mfj"},
		},
		{
			name:          "HTTP error returns error",
			spec:          Spec{Ecosystem: "PyPI", Name: "jinja2"},
			responses:     map[string]string{"3.1.2|": `{"code": 3, "message": "Invalid query."}`},
			statusCode:    400,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := New(tt.spec)
			require.NoError(t, err)

			mock := &mockOSV{responses: tt.responses, statusCode: tt.statusCode}
			o.webClient = mock.client()

			groups, err := o.knownVulnerabilities(context.Background(), "3.1.2")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			ids := []string{}
			for _, group := range groups {
				ids = append(ids, group.ID)
			}
			assert.Equal(t, tt.expectedIDs, ids)
		})
	}
}

func TestVulnGroup(t *testing.T) {
	o, err := New(Spec{Ecosystem: "PyPI", Name: "Jinja2"})
	require.NoError(t, err)

	mock := &mockOSV{responses: map[string]string{"3.1.2|": jinja2At312}}
	o.webClient = mock.client()

	groups, err := o.knownVulnerabilities(context.Background(), "3.1.2")
	require.NoError(t, err)
	require.NotEmpty(t, groups)

	// The GHSA severity is kept although the PYSEC twin has none,
	// and fixed versions are found although the package name case differs.
	assert.Equal(t,
		"GHSA-cpwx-vrp4-4pq7 (CVE-2025-27516, PYSEC-2026-1471) [MODERATE] Summary of GHSA-cpwx-vrp4-4pq7 - fixed in: 3.1.6",
		groups[0].String())
}
