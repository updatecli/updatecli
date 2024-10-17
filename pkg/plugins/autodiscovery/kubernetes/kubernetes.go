package kubernetes

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func (k Kubernetes) discoverKubernetesManifest(file, relativeFile string) [][]byte {
	var manifests [][]byte

	data, err := getKubernetesManifestData(file)
	if err != nil {
		logrus.Debugln(err)
		return manifests
	}
	if data == nil {
		return manifests
	}
	for i, container := range data.Spec.Containers {

		containerName := container.Name
		if containerName == "" {
			containerName = container.Image
		}

		manifest, err := k.generateContainerManifest(
			fmt.Sprintf("$.spec.containers[%d].image", i),
			containerName,
			container.Image,
			relativeFile,
			"")

		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if manifest == nil {
			continue
		}

		manifests = append(manifests, manifest)
	}

	for i, container := range data.Spec.Template.Spec.Containers {

		containerName := container.Name
		if containerName == "" {
			containerName = container.Image
		}

		manifest, err := k.generateContainerManifest(
			fmt.Sprintf("$.spec.template.spec.containers[%d].image", i),
			containerName,
			container.Image,
			relativeFile,
			"")

		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if manifest == nil {
			continue
		}

		manifests = append(manifests, manifest)
	}
	return manifests
}
