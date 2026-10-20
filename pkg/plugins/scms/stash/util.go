package stash

import (
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"
)

func (s *Stash) setDirectory() {

	if _, err := os.Stat(s.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(s.Spec.Directory, 0755)
		if err != nil {
			log.Errorf("err - %s", err)
		}
	}
}
