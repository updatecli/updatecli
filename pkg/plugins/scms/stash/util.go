package stash

import (
	"os"

	"github.com/sirupsen/logrus"
)

func (s *Stash) setDirectory() {

	if _, err := os.Stat(s.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(s.Spec.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}
