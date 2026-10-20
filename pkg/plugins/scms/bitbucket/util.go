package bitbucket

import (
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"
)

func (b *Bitbucket) setDirectory() {
	if _, err := os.Stat(b.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(b.Spec.Directory, 0o755)
		if err != nil {
			log.Errorf("err - %s", err)
		}
	}
}
