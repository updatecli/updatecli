package registry

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	var err error

	if fileStore == "" {
		fileStore, err = os.Getwd()
		if err != nil {
			logrus.Errorln(err)
		}
	}
}
