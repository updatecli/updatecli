package registry

var (
	// updatecliManifestMediaType is the OCI media type for updatecli manifests.
	updatecliManifestMediaType string = "application/io.updatecli.policy.manifest"
	// updatecliValueMediaType is the OCI media type for updatecli value file.
	updatecliValueMediaType string = "application/io.updatecli.policy.value"
	// updatecliSecretMediaType is the OCI media type for updatecli secret file.
	updatecliSecretMediaType string = "application/io.updatecli.policy.secret"
	// ociArtifactType is the media type for updatecli OCI artifacts.
	ociArtifactType string = "application/io.updatecli.policy"
	// ociDefaultTag is the default tag for updatecli OCI images.
	ociDefaultTag string = "latest"
)
