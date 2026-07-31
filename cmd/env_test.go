package cmd

import (
	"testing"
)

func TestGetEnvBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		setEnv       bool
		defaultValue bool
		expected     bool
	}{
		{
			name:         "not_set_returns_default_true",
			envVar:       "UNDEFINED_VAR_1",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "not_set_returns_default_false",
			envVar:       "UNDEFINED_VAR_2",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "set_to_true",
			envVar:       "TEST_VAR_TRUE",
			envValue:     "true",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "set_to_false",
			envVar:       "TEST_VAR_FALSE",
			envValue:     "false",
			setEnv:       true,
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "set_to_1",
			envVar:       "TEST_VAR_1",
			envValue:     "1",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "set_to_0",
			envVar:       "TEST_VAR_0",
			envValue:     "0",
			setEnv:       true,
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "whitespace_trimmed",
			envVar:       "TEST_VAR_SPACE",
			envValue:     "  true  ",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "invalid_returns_default_true",
			envVar:       "TEST_VAR_INVALID_1",
			envValue:     "invalid",
			setEnv:       true,
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "invalid_returns_default_false",
			envVar:       "TEST_VAR_INVALID_2",
			envValue:     "maybe",
			setEnv:       true,
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envVar, tt.envValue)
			}

			result := getEnvBoolOrDefault(tt.envVar, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
