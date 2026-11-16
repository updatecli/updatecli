package githubaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

func TestDetectVersionFilter(t *testing.T) {
	allowDowngrade := true
	cooldown := age.AutodiscoverySpec{Spec: age.Spec{Minimum: "7d"}}
	cooldownAllowingDowngrade := age.AutodiscoverySpec{Spec: age.Spec{Minimum: "7d"}, AllowDowngrade: &allowDowngrade}

	tests := []struct {
		name        string
		reference   string
		releaseAge  age.AutodiscoverySpec
		wantKind    string
		wantPattern string
	}{
		{name: "semver reference bounded at itself", reference: "v4.2.1", releaseAge: cooldown, wantKind: "semver", wantPattern: ">=4.2.1"},
		{name: "semver reference accepts any version when a downgrade is allowed", reference: "v4.2.1", releaseAge: cooldownAllowingDowngrade, wantKind: "semver", wantPattern: "*"},
		{name: "prerelease reference keeps accepting prereleases when a downgrade is allowed", reference: "v2.0.0-rc.2", releaseAge: cooldownAllowingDowngrade, wantKind: "semver", wantPattern: ">=0.0.0-0"},
		{name: "prerelease reference keeps accepting newer prereleases", reference: "v2.0.0-rc.2", releaseAge: cooldown, wantKind: "semver", wantPattern: ">=2.0.0-rc.2"},
		// Without an age filter, nothing holds back the newest versions, so the pattern is left as it was
		{name: "semver reference without age filter", reference: "v4.2.1", wantKind: "semver", wantPattern: "*"},
		{name: "branch reference", reference: "main", releaseAge: cooldown, wantKind: latestVersionIdentifier, wantPattern: latestVersionIdentifier},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, pattern := detectVersionFilter(tt.reference, tt.releaseAge)
			assert.Equal(t, tt.wantKind, kind)
			assert.Equal(t, tt.wantPattern, pattern)
		})
	}
}
