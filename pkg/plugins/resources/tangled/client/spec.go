package client

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "tangled" resource
// parsed from an updatecli manifest file.
type Spec struct {
	// "appview" defines the Tangled appview URL used for user-facing pull request links.
	//
	// default:
	//   https://tangled.sh
	Appview string `yaml:",omitempty"`
	// "pds" defines the URL of the user's atproto Personal Data Server (PDS).
	//
	// default:
	//   https://bsky.social
	//
	// remark:
	//   If left empty, the plugin will fall back to "https://bsky.social".
	//   For self-hosted PDS instances or atproto PDSes other than Bluesky, the
	//   full URL (including scheme) must be provided.
	PDS string `yaml:",omitempty"`
	// "did" specifies the user's DID (decentralized identifier) used to author pull request records.
	//
	// remark:
	//   either "did" or "handle" must be set. "did" takes precedence.
	DID string `yaml:",omitempty"`
	// "handle" specifies the user's atproto handle (e.g. alice.tangled.sh).
	//
	// remark:
	//   either "did" or "handle" must be set. When only "handle" is set, the
	//   plugin will resolve the handle to a DID using the configured PDS.
	Handle string `yaml:",omitempty"`
	// "appPassword" specifies the atproto app password used to authenticate the user.
	//
	// remark:
	//   App passwords are sensitive and should not be stored in plaintext within
	//   the manifest. Use environment variables or a SOPS file instead, for
	//   example: `{{ requiredEnv "TANGLED_APP_PASSWORD" }}`.
	AppPassword string `yaml:",omitempty"`
}

// Validate validates that a spec contains the credentials required to
// authenticate against the user's PDS. Callers that only need git transport
// (clone/push over SSH) should skip this check — auth is only enforced when
// an atproto session is actually opened.
func (s Spec) Validate() error {
	if s.DID == "" && s.Handle == "" {
		logrus.Errorf("missing %q or %q parameter", "did", "handle")
		return fmt.Errorf("wrong tangled configuration")
	}

	if s.AppPassword == "" {
		logrus.Errorf("missing %q parameter", "appPassword")
		return fmt.Errorf("wrong tangled configuration")
	}

	return nil
}

// Sanitize applies default URL values and normalizes URL schemes. It does
// not validate atproto credentials — that happens lazily when the client
// actually needs to log in.
func (s *Spec) Sanitize() {
	if s.Appview == "" {
		s.Appview = "https://tangled.sh"
	}

	if s.PDS == "" {
		s.PDS = "https://bsky.social"
	}

	s.Appview = normalizeURL(s.Appview)
	s.PDS = normalizeURL(s.PDS)
}

func normalizeURL(u string) string {
	u = strings.TrimRight(u, "/")
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		u = "https://" + u
	}
	return u
}
