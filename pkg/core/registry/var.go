package registry

var (
	// updatecliManifestMediaType is the OCI media type for updatecli manifests.
	updatecliManifestMediaType string = "application/io.updatecli.policy.manifest.alpha"
	// updatecliValueMediaType is the OCI media type for updatecli value file.
	updatecliValueMediaType string = "application/io.updatecli.policy.value.alpha"
	// updatecliSecretMediaType is the OCI media type for updatecli secret file.
	updatecliSecretMediaType string = "application/io.updatecli.policy.secret.alpha"
	// ociArtifactType is the media type for updatecli OCI artifacts.
	ociArtifactType string = "application/io.updatecli.policy.alpha"
	// ociDefaultTag is the default tag for updatecli OCI images.
	ociDefaultTag string = "latest"
)
