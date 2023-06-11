package dockerdigest

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// Spec defines a specification for a "dockerdigest" resource parsed from an updatecli manifest file
type Spec struct {
	// [s][c] Architecture specifies the container image architecture such as `amd64`
	Architecture string `yaml:",omitempty"`
	// [s][c] Image specifies the container image such as `updatecli/updatecli`
	Image string `yaml:",omitempty"`
	// [s] Tag specifies the container image tag such as `latest`
	Tag string `yaml:",omitempty"`
	// [c] Digest specifies the container image digest such as `@sha256:ce782db15ab5491c6c6178da8431b3db66988ccd11512034946a9667846952a6`
	Digest                string `yaml:",omitempty"`
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Architecture: s.Architecture,
		Image:        s.Image,
		Tag:          s.Tag,
		Digest:       s.Digest,
	}
}
