package kubernetes

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

func (k Kubernetes) discoverProwManifest(file, relativeFile string) [][]byte {
	var manifests [][]byte

	prowManifestData, err := getProwManifestData(file)
	if err != nil {
		logrus.Debugln(err)
		return manifests
	}
	if prowManifestData == nil {
		return manifests
	}
	// Prow Presubmit
	for document, data := range prowManifestData {
		for repo, tests := range data.ProwPreSubmitJobs {
			for i, test := range tests {
				for j, container := range test.Spec.Containers {
					containerName := container.Name
					if containerName == "" {
						imageName, _, _, err := dockerimage.ParseOCIReferenceInfo(container.Image)
						if err != nil {
							logrus.Debugln(err)
							continue
						}
						containerName = imageName
					}
					manifest, err := k.generateContainerManifest(
						fmt.Sprintf("$.presubmits.'%s'[%d].spec.containers[%d].image", repo, i, j),
						containerName,
						container.Image,
						relativeFile,
						fmt.Sprintf(" for repo %q and presubmit test %q", repo, test.Name),
						document,
					)

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
		}
	}

	// Prow Postsubmit
	for document, d := range prowManifestData {
		for repo, tests := range d.ProwPostSubmitJobs {
			for i, test := range tests {
				for j, container := range test.Spec.Containers {
					containerName := container.Name
					if containerName == "" {
						imageName, _, _, err := dockerimage.ParseOCIReferenceInfo(container.Image)
						if err != nil {
							logrus.Debugln(err)
							continue
						}
						containerName = imageName
					}
					manifest, err := k.generateContainerManifest(
						fmt.Sprintf("$.postsubmits.'%s'[%d].spec.containers[%d].image", repo, i, j),
						containerName,
						container.Image,
						relativeFile,
						fmt.Sprintf(" for repo %q and postsubmit test %q", repo, test.Name),
						document,
					)

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
		}
	}

	// Prow Periodics
	for document, d := range prowManifestData {
		for i, test := range d.ProwPeriodicJobs {
			for j, container := range test.Spec.Containers {
				containerName := container.Name
				if containerName == "" {

					imageName, _, _, err := dockerimage.ParseOCIReferenceInfo(container.Image)
					if err != nil {
						logrus.Debugln(err)
						continue
					}
					containerName = imageName
				}
				manifest, err := k.generateContainerManifest(
					fmt.Sprintf("$.periodics[%d].spec.containers[%d].image", i, j),
					containerName,
					container.Image,
					relativeFile,
					fmt.Sprintf(" for periodic test %q", test.Name),
					document,
				)

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
	}
	return manifests
}
