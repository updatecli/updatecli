package pullrequest

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"

	"github.com/updatecli/updatecli/pkg/plugins/resources/tangled/client"
	tangledscm "github.com/updatecli/updatecli/pkg/plugins/scms/tangled"
)

// Tangled contains information to interact with a Tangled pullrequest.
type Tangled struct {
	spec   Spec
	client *client.Client
	scm    *tangledscm.Tangled

	SourceBranch string
	TargetBranch string
	Knot         string
	Owner        string
	Repository   string
}

// New returns a new valid Tangled object.
func New(spec any, scm *tangledscm.Tangled) (Tangled, error) {
	var clientSpec client.Spec
	var s Spec

	if err := mapstructure.Decode(spec, &clientSpec); err != nil {
		return Tangled{}, fmt.Errorf("error decoding client spec: %w", err)
	}

	if err := mapstructure.Decode(spec, &s); err != nil {
		return Tangled{}, fmt.Errorf("error decoding spec: %w", err)
	}

	prOverridesClient := clientSpecOverridden(clientSpec)

	if scm != nil {
		if clientSpec.AppPassword == "" {
			clientSpec.AppPassword = scm.Spec.AppPassword
		}
		if clientSpec.DID == "" {
			clientSpec.DID = scm.Spec.DID
		}
		if clientSpec.Handle == "" {
			clientSpec.Handle = scm.Spec.Handle
		}
		if clientSpec.PDS == "" {
			clientSpec.PDS = scm.Spec.PDS
		}
		if clientSpec.Appview == "" {
			clientSpec.Appview = scm.Spec.Appview
		}
	}

	clientSpec.Sanitize()

	// CreateAction will hit the PDS so credentials must be present by the
	// time we get here. Validate eagerly to surface misconfiguration before
	// any pipeline work runs.
	if err := clientSpec.Validate(); err != nil {
		return Tangled{}, err
	}

	// Reuse the SCM session only when the manifest did not supply
	// PR-specific overrides — otherwise a separate atproto session is needed
	// to honor the alternate identity/PDS/appview.
	var c *client.Client
	if !prOverridesClient && scm != nil && scm.Client() != nil {
		c = scm.Client()
	} else {
		var err error
		c, err = client.New(clientSpec)
		if err != nil {
			return Tangled{}, err
		}
	}

	t := Tangled{
		spec:   s,
		client: c,
		scm:    scm,
	}

	t.inheritFromScm()

	return t, nil
}

// clientSpecOverridden reports whether the PR spec carries explicit auth or
// appview values that should be honored instead of inheriting the SCM session.
func clientSpecOverridden(s client.Spec) bool {
	return s.DID != "" || s.Handle != "" || s.AppPassword != "" || s.PDS != "" || s.Appview != ""
}
