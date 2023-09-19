package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
)

func GenerateSchema(baseSchemaID, schemaDir string) error {

	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Json Schema")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Json Schema"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Json Schema")+4))

	err := jsonschema.CloneCommentDirectory()

	if err != nil {
		return err
	}

	defer func() {
		tmperr := jsonschema.CleanCommentDirectory()
		if err != nil {
			err = fmt.Errorf("%s\n%s", err, tmperr)
		}
	}()

	s := jsonschema.New(baseSchemaID, schemaDir)
	err = s.GenerateSchema(&config.Spec{})
	if err != nil {
		return err
	}

	logrus.Infof("```\n%s\n```\n", s)

	err = s.Save()
	if err != nil {
		return err
	}

	return s.GenerateSchema(&config.Spec{})
}
