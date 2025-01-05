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

	list := []struct {
		cs  []containerSpec
		key string
	}{
		{
			cs:  data.Spec.Containers,
			key: "$.spec.containers[%d].image",
		},
		{
			cs:  data.Spec.Template.Spec.Containers,
			key: "$.spec.template.spec.containers[%d].image",
		},
		{
			cs:  data.Spec.JobTemplateSpec.Spec.Template.Spec.Containers,
			key: "$.spec.jobTemplate.spec.template.spec.containers[%d].image",
		},
	}
	for _, v := range list {
		for i, container := range v.cs {
			containerName := container.Name
			if containerName == "" {
				containerName = container.Image
			}

			manifest, err := k.generateContainerManifest(
				fmt.Sprintf(v.key, i),
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
	}
	return manifests
}
