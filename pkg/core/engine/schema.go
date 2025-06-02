package engine

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/compose"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/registry"
)

func GenerateSchema(baseSchemaID, schemaDir string) error {
	PrintTitle("Json Schema")

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

	generateSchema := func(baseSchemaID, schemaDir, subSchemaDir string, spec interface{}) error {
		if subSchemaDir != "" {
			schemaDir = filepath.Join(schemaDir, subSchemaDir)

			u, err := url.Parse(baseSchemaID)
			if err != nil {
				return err
			}

			u.Path = filepath.Join(u.Path, subSchemaDir)

			baseSchemaID = u.String()
		}

		s := jsonschema.New(baseSchemaID, schemaDir)
		err = s.GenerateSchema(spec)
		if err != nil {
			return err
		}

		logrus.Infof("```\n%s\n```\n", s)

		err = s.Save()
		if err != nil {
			return fmt.Errorf("unable to save schema - %s", err)
		}

		return nil
	}

	if err = generateSchema(baseSchemaID, schemaDir, "policy/manifest", config.Spec{}); err != nil {
		return fmt.Errorf("unable to generate schema - %s", err)
	}

	if err = generateSchema(baseSchemaID, schemaDir, "policy/metadata", registry.PolicySpec{}); err != nil {
		return fmt.Errorf("unable to generate schema - %s", err)
	}

	if err = generateSchema(baseSchemaID, schemaDir, "compose", compose.Spec{}); err != nil {
		return fmt.Errorf("unable to generate schema - %s", err)
	}

	return nil
}
