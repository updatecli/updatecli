package gitlab

import (
	"os"

	"github.com/sirupsen/logrus"
)

// setDirectory creates the local git repository path if it does not exist.
func (g *Gitlab) setDirectory() {

	if _, err := os.Stat(g.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Spec.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}
