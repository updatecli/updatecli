package gitea

import (
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"
)

func (g *Gitea) setDirectory() {

	if _, err := os.Stat(g.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Spec.Directory, 0755)
		if err != nil {
			log.Errorf("err - %s", err)
		}
	}
}
