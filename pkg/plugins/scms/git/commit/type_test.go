package commit

import (
	"strings"
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"go.yaml.in/yaml/v4"
)

func TestCommitTypeConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, config, prefix string
	}{
		{"unset", "{}", "chore: "},
		{"unset with scope", "scope: deps", "chore(deps): "},
		{"empty", "type: \"\"", ""},
		{"empty with scope", "type: \"\"\nscope: deps", "deps: "},
		{"explicit", "type: fix", "fix: "},
		{"explicit with scope", "type: fix\nscope: deps", "fix(deps): "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, decode := range []string{"yaml", "mapstructure"} {
				t.Run(decode, func(t *testing.T) {
					var c Commit
					if decode == "yaml" {
						if err := yaml.Unmarshal([]byte(tc.config), &c); err != nil {
							t.Fatal(err)
						}
					} else {
						var spec map[string]any
						if err := yaml.Unmarshal([]byte(tc.config), &spec); err != nil {
							t.Fatal(err)
						}
						if err := mapstructure.Decode(spec, &c); err != nil {
							t.Fatal(err)
						}
					}

					// Autodiscovery serializes SCM settings into generated manifests.
					// In particular, omitempty must preserve an explicit empty type.
					encoded, err := yaml.Marshal(c)
					if err != nil {
						t.Fatal(err)
					}
					var restored Commit
					if err := yaml.Unmarshal(encoded, &restored); err != nil {
						t.Fatal(err)
					}
					restored.HideCredit = true
					// Generating multiple titles also revalidates the same configuration.
					for _, title := range []string{"Bump dependency", strings.Repeat("a", 72-len(tc.prefix))} {
						got, err := restored.Generate(title)
						if err != nil {
							t.Fatal(err)
						}
						if want := tc.prefix + title; got != want {
							t.Errorf("Generate() = %q, want %q", got, want)
						}
					}
					got, err := restored.Generate(strings.Repeat("a", 73-len(tc.prefix)))
					if err != nil {
						t.Fatal(err)
					}
					if title := strings.SplitN(got, "\n", 2)[0]; len(title) != 72 || !strings.HasSuffix(title, "...") {
						t.Errorf("long title = %q, want 72 characters ending in ...", title)
					}
				})
			}
		})
	}
}
