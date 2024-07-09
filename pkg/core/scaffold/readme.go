package scaffold

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	readmeFile     string = "README.md"
	readmeTemplate string = `# README

## REQUIREMENTS

Before using this policy, here is a list of requirements to meet:

- [ ] Review the file [Policy.yaml](./Policy.yaml) and update the values to match your needs (see [documentation](https://www.updatecli.io/docs/core/shareandreuse/) section).
- [ ] Review the file [values.d/default.yaml](./values.d/default.yaml) and update the values to match your needs.
- [ ] Review the file [updatecli.d/default.yaml](./updatecli.d/default.yaml) and update the values to match your needs (see [documentation](https://www.updatecli.io/docs/core/configuration/) section).

Feel free to delete this requirements section once you are done with the checklist.

## DESCRIPTION

<Please specify your policy description >


## HOW TO USE

**Show**

They are two different approaches to see the content of this policy:

Using the policy from the local filesystem by running:

	updatecli manifest show --config updatecli.d --values values.d/default.yaml

Using the policy from the registry by running:

    updatecli manifest show $OCI_REGISTRY/< insert your policy name>:v0.1.0


**Use**

Similarly to the show command, they are two ways to execute an Updatecli policy, either using the local file or the one stored on the registry.

Using the policy from the local filesystem by running:

    updatecli diff --config updatecli.d --values values.d/default.yaml

Using the policy from the registry by running:

    updatecli diff ghcr.io/updatecli/policies/<a policy name>:v0.1.0


If "diff" is replaced by "apply", then the policy will be executed in enforce mode.

⚠ Any values files specified at runtime will override default values set from the policy bundle

**Login**

Regardless your Updatecli policy is meant to be public or private, you probably always want to be authenticated with your registry, by running:

    docker login "$OCI_REGISTRY"

INFO: OCI_REGISTRY can be any OCI compliant registry such as [Zot](https://github.com/project-zot/zot), [DockerHub](https://hub.docker.com), [ghcr.io](https://ghcr.io),etc.

**Publish**

Policies defines in this repository can be published to your registry by running:

	updatecli manifest push \
		--config updatecli.d \
		--values values.d/default.yaml \
    	--policy Policy.yaml \
    	--tag "$OCI_REGISTRY/<insert your policy name>" \
		.

⚠ The tag is defined by the version field in the policy file
⚠ The latest tag always represents the latest version published from
a semantic versioning point of view.

## NEXT STEPS

Feel free to look on the [Updatecli documentation](https://updatecli.io) to learn more about how to use Updatecli.

Another good starting point is to understand how to use [updatecli-compose.yaml](https://www.updatecli.io/docs/core/compose/) to orchestrate multiple Updatecli policies.

## CONTRIBUTING

This document has been generated from this [template](https://github.com/updatecli/updatecli/blob/main/pkg/core/scaffold/readme.go).
Feel free to suggest any improvements or open an [issue](https://github.com/updatecli/updatecli/issues).
`
)

func (s *Scaffold) scaffoldReadme(dirname string) error {

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	readmeFilePath := filepath.Join(dirname, readmeFile)

	// If the config already exist, we don't overwrite it
	if _, err := os.Stat(readmeFilePath); err == nil {
		logrus.Infof("Skipping, readme already exist: %s", readmeFilePath)
		return nil
	}

	f, err := os.Create(readmeFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(readmeTemplate))
	if err != nil {
		return err
	}

	return nil
}
