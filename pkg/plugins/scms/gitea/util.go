package gitea

import (
	"os"

	"github.com/sirupsen/logrus"
)

func (g *Gitea) setDirectory() {

	if _, err := os.Stat(g.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Spec.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}
