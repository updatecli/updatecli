package githubaction

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

// dockerGHAManifestSpec contains the parameters to generate the Updatecli manifest specifically for Docker images.
type dockerGHAManifestSpec struct {
	ActionName        string
	Image             string
	RelativeFoundFile string
	TargetKey         string
}

// getDockerManifest returns the Updatecli manifest based on the Docker image
func (g GitHubAction) getDockerManifest(spec *dockerGHAManifestSpec) ([]byte, error) {

	if spec.Image == "" {
		return nil, fmt.Errorf("image is empty")
	}

	imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(strings.TrimPrefix(spec.Image, "docker://"))
	if err != nil {
		return nil, fmt.Errorf("parsing image %q: %s", spec.Image, err)
	}

	/*
		For the time being, it's not possible to retrieve a list of tag for a specific digest
		without a significant amount f api call. More information on following issue
		https://github.com/google/go-containerregistry/issues/1297
		until a better solution, we don't handle docker image digest
	*/
	if imageDigest != "" && imageTag == "" {
		log.Debugf("docker digest without specified tag is not supported at the moment for %q", spec.Image)
		return nil, nil
	}

	// Test if the ignore rule based on path is respected
	if len(g.spec.Ignore) > 0 {
		if g.spec.Ignore.isMatchingRules(
			g.rootDir,
			spec.RelativeFoundFile,
			spec.Image,
			imageTag) {

			log.Debugf("Ignoring Docker image %q from %q, as not matching ignore rule(s)\n",
				spec.Image,
				spec.RelativeFoundFile)
			return nil, nil
		}
	}

	// Test if the only rule based on path is respected
	if len(g.spec.Only) > 0 {
		if !g.spec.Only.isMatchingRules(
			g.rootDir,
			spec.RelativeFoundFile,
			spec.Image,
			imageTag) {

			log.Debugf("Ignoring Docker image %q from %q, as not matching only rule(s)\n",
				spec.Image,
				spec.RelativeFoundFile)
			return nil, nil
		}
	}

	// New versions are looked up under the upstream name, while the target keeps
	// the image name written in the file.
	sourceImageName := g.spec.RegistryMappings.Resolve(imageName)

	sourceSpec := dockerimage.NewDockerImageSpecFromImage(sourceImageName, imageTag, g.spec.CredentialsDocker)

	versionFilterKind := g.versionFilter.Kind
	versionFilterPattern := g.versionFilter.Pattern
	versionFilterRegex := g.versionFilter.Regex
	tagFilter := "*"
	architecture := ""

	if sourceSpec != nil {
		versionFilterKind = sourceSpec.VersionFilter.Kind
		versionFilterPattern = sourceSpec.VersionFilter.Pattern
		versionFilterRegex = sourceSpec.VersionFilter.Regex
		tagFilter = sourceSpec.TagFilter
		architecture = sourceSpec.Architecture
	}

	// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence.
	// The tag filter derived from the current tag is kept, so that the image variant such as "alpine"
	// doesn't change: the autodiscovery spec has no tag filter to override it.
	if !g.spec.VersionFilter.IsZero() {
		versionFilterKind = g.versionFilter.Kind
		versionFilterPattern, err = g.versionFilter.GreaterThanPattern(imageTag)
		versionFilterRegex = g.versionFilter.Regex
		if err != nil {
			log.Debugf("building version filter pattern: %s", err)
			versionFilterPattern = "*"
		}
	}

	var tmpl *template.Template
	if g.digest && sourceSpec != nil {
		tmpl, err = template.New("manifest").Parse(manifestTemplateDockerDigestAndLatest)
		if err != nil {
			return nil, err
		}
	} else if g.digest && sourceSpec == nil {
		tmpl, err = template.New("manifest").Parse(manifestTemplateDockerDigest)
		if err != nil {
			return nil, err
		}
	} else if !g.digest && sourceSpec != nil {
		tmpl, err = template.New("manifest").Parse(manifestTemplateDockerLatest)
		if err != nil {
			return nil, err
		}
	} else {
		log.Infoln("No source spec detected")
		return nil, nil
	}

	// A mirror may not serve a version found upstream yet, or may serve it under
	// another digest, so the target only runs once the mirror has what it writes.
	// Every template writes the "digest" source with digest enabled, else the "image" one.
	mirrorConditionSourceID := "image"
	if g.digest {
		mirrorConditionSourceID = "digest"
	}
	mirrorCredential, _ := dockerimage.RegistryCredentials(imageName, g.spec.CredentialsDocker)

	targetPrefix := imageName + ":"
	if strings.HasPrefix(spec.Image, "docker://") {
		targetPrefix = "docker://" + imageName + ":"
	}

	params := struct {
		ActionName              string
		ActionID                string
		ImageName               string
		SourceImageName         string
		ImageTag                string
		ImageArchitecture       string
		TargetName              string
		TargetFile              string
		TargetKey               string
		TargetPrefix            string
		TagFilter               string
		VersionFilterKind       string
		VersionFilterPattern    string
		VersionFilterRegex      string
		ScmID                   string
		MirrorCondition         bool
		MirrorConditionSourceID string
		MirrorUsername          string
		MirrorPassword          string
		MirrorToken             string
	}{
		ActionName:              spec.ActionName,
		ActionID:                g.actionID,
		ImageName:               imageName,
		SourceImageName:         sourceImageName,
		ImageTag:                imageTag,
		ImageArchitecture:       architecture,
		TargetName:              fmt.Sprintf(`deps: bump docker image %q in %q to {{ source %q }}`, imageName, spec.RelativeFoundFile, imageName),
		TargetFile:              spec.RelativeFoundFile,
		TargetKey:               spec.TargetKey,
		TargetPrefix:            targetPrefix,
		TagFilter:               tagFilter,
		VersionFilterKind:       versionFilterKind,
		VersionFilterPattern:    versionFilterPattern,
		VersionFilterRegex:      versionFilterRegex,
		ScmID:                   g.scmID,
		MirrorCondition:         sourceImageName != imageName,
		MirrorConditionSourceID: mirrorConditionSourceID,
		MirrorUsername:          mirrorCredential.Username,
		MirrorPassword:          mirrorCredential.Password,
		MirrorToken:             mirrorCredential.Token,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}
