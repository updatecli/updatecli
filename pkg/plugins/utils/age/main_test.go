package age

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReleaseAge(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "hours via standard Go duration",
			input:    "24h",
			expected: 24 * time.Hour,
		},
		{
			name:     "days suffix",
			input:    "1d",
			expected: 24 * time.Hour,
		},
		{
			name:     "multiple days",
			input:    "7d",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "weeks suffix",
			input:    "1w",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "multiple weeks",
			input:    "2w",
			expected: 14 * 24 * time.Hour,
		},
		{
			name:     "months suffix",
			input:    "1mo",
			expected: time.Duration(24*365/12) * time.Hour,
		},
		{
			name:     "multiple months",
			input:    "2mo",
			expected: time.Duration(2*24*365/12) * time.Hour,
		},
		{
			name:     "years suffix",
			input:    "1y",
			expected: 365 * 24 * time.Hour,
		},
		{
			name:     "multiple years",
			input:    "2y",
			expected: 2 * 365 * 24 * time.Hour,
		},
		{
			name:     "minutes via standard Go duration",
			input:    "30m",
			expected: 30 * time.Minute,
		},
		{
			name:     "leading and trailing spaces are trimmed",
			input:    "  7d  ",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:    "invalid value",
			input:   "notaduration",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseReleaseAge(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSpecValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name:    "empty spec is valid",
			spec:    Spec{},
			wantErr: false,
		},
		{
			name:    "valid minimum only",
			spec:    Spec{Minimum: "7d"},
			wantErr: false,
		},
		{
			name:    "valid maximum only",
			spec:    Spec{Maximum: "30d"},
			wantErr: false,
		},
		{
			name:    "valid minimum and maximum",
			spec:    Spec{Minimum: "7d", Maximum: "30d"},
			wantErr: false,
		},
		{
			name:    "valid minimum with hours",
			spec:    Spec{Minimum: "48h"},
			wantErr: false,
		},
		{
			name:    "valid minimum with weeks",
			spec:    Spec{Minimum: "2w"},
			wantErr: false,
		},
		{
			name:    "valid minimum with months",
			spec:    Spec{Minimum: "1mo"},
			wantErr: false,
		},
		{
			name:    "valid minimum with years",
			spec:    Spec{Minimum: "1y"},
			wantErr: false,
		},
		{
			name:    "invalid minimum",
			spec:    Spec{Minimum: "notaduration"},
			wantErr: true,
		},
		{
			name:    "invalid maximum",
			spec:    Spec{Maximum: "notaduration"},
			wantErr: true,
		},
		{
			name:    "invalid maximum with valid minimum",
			spec:    Spec{Minimum: "7d", Maximum: "notaduration"},
			wantErr: true,
		},
		{
			name:    "invalid minimum with valid maximum",
			spec:    Spec{Minimum: "notaduration", Maximum: "30d"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSpecIsOlderThan(t *testing.T) {
	// Use a fixed reference time so tests are deterministic.
	since := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		spec        Spec
		releaseTime time.Time
		want        bool
	}{
		{
			name:        "no minimum set always returns false",
			spec:        Spec{},
			releaseTime: since.Add(-3 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "release is younger than minimum: should be skipped",
			spec:        Spec{Minimum: "7d"},
			releaseTime: since.Add(-3 * 24 * time.Hour),
			want:        true,
		},
		{
			name:        "release is older than minimum: should be included",
			spec:        Spec{Minimum: "7d"},
			releaseTime: since.Add(-10 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "release exactly at minimum boundary: should be included",
			spec:        Spec{Minimum: "7d"},
			releaseTime: since.Add(-7 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "minimum expressed in hours",
			spec:        Spec{Minimum: "48h"},
			releaseTime: since.Add(-24 * time.Hour),
			want:        true,
		},
		{
			name:        "minimum expressed in weeks",
			spec:        Spec{Minimum: "1w"},
			releaseTime: since.Add(-10 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "release is in the future relative to since",
			spec:        Spec{Minimum: "7d"},
			releaseTime: since.Add(24 * time.Hour),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.IsOlderThan(tt.releaseTime, &since)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSpecIsNewerThan(t *testing.T) {
	// Use a fixed reference time so tests are deterministic.
	since := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		spec        Spec
		releaseTime time.Time
		want        bool
	}{
		{
			name:        "no maximum set always returns false",
			spec:        Spec{},
			releaseTime: since.Add(-40 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "release is older than maximum: should be skipped",
			spec:        Spec{Maximum: "30d"},
			releaseTime: since.Add(-40 * 24 * time.Hour),
			want:        true,
		},
		{
			name:        "release is within maximum: should be included",
			spec:        Spec{Maximum: "30d"},
			releaseTime: since.Add(-10 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "release exactly at maximum boundary: should be included",
			spec:        Spec{Maximum: "30d"},
			releaseTime: since.Add(-30 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "maximum expressed in hours",
			spec:        Spec{Maximum: "48h"},
			releaseTime: since.Add(-72 * time.Hour),
			want:        true,
		},
		{
			name:        "maximum expressed in weeks",
			spec:        Spec{Maximum: "1w"},
			releaseTime: since.Add(-3 * 24 * time.Hour),
			want:        false,
		},
		{
			name:        "release is in the future relative to since",
			spec:        Spec{Maximum: "30d"},
			releaseTime: since.Add(24 * time.Hour),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.IsNewerThan(tt.releaseTime, &since)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSpecIsZero(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want bool
	}{
		{
			name: "empty spec is zero",
			spec: Spec{},
			want: true,
		},
		{
			name: "minimum set is not zero",
			spec: Spec{Minimum: "7d"},
			want: false,
		},
		{
			name: "maximum set is not zero",
			spec: Spec{Maximum: "30d"},
			want: false,
		},
		{
			name: "both set is not zero",
			spec: Spec{Minimum: "7d", Maximum: "30d"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.IsZero()
			assert.Equal(t, tt.want, got)
		})
	}
}
