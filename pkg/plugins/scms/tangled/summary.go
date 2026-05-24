package tangled

import "fmt"

// Summary returns a brief description of the Tangled SCM configuration.
func (t *Tangled) Summary() string {
	return fmt.Sprintf("%s/%s/%s@%s", t.Spec.Knot, t.Spec.Owner, t.Spec.Repository, t.Spec.Branch)
}
