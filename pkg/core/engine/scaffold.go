package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/scaffold"
)

func (e *Engine) Scaffold(rootDir string) error {

	PrintTitle("Scaffold a new Updatecli policy")

	s := scaffold.Scaffold{}

	if err := s.Run(rootDir); err != nil {
		return fmt.Errorf("unable to scaffold a new Updatecli policy - %s", err)
	}

	logrus.Infof("A new Updatecli policy has been scaffolded in %s", rootDir)

	return nil
}
