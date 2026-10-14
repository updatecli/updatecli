package gitlab

import (
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"
)

// setDirectory creates the local git repository path if it does not exist.
func (g *Gitlab) setDirectory() {

	if _, err := os.Stat(g.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Spec.Directory, 0755)
		if err != nil {
			log.Errorf("err - %s", err)
		}
	}
}
