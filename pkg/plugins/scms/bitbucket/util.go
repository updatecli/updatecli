package bitbucket

import (
	"os"

	"github.com/sirupsen/logrus"
)

func (b *Bitbucket) setDirectory() {
	if _, err := os.Stat(b.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(b.Spec.Directory, 0o755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}
