package bitbucket

import (
	"fmt"
)

// Summary returns a brief description of the Bitbucket SCM configuration
func (b *Bitbucket) Summary() string {

	return fmt.Sprintf("bitbucket.org/%s/%s@%s", b.Spec.Owner, b.Spec.Repository, b.Spec.Branch)
}
