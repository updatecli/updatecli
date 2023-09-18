package registry

var (
	// updatecliManifestMediaType is the OCI media type for updatecli manifests.
	updatecliManifestMediaType string = "application/io.updatecli.manifest"
	// updatecliValueMediaType is the OCI media type for updatecli value file.
	updatecliValueMediaType string = "application/io.updatecli.value"
	// updatecliSecretMediatType is the OCI media type for updatecli secret file.
	updatecliSecretMediaType string = "application/io.updatecli.secret"
	// ociArtifactType is the media type for updatecli OCI artifacts.
	ociArtifactType string = "application/io.updatecli.artifact"
	// ociDefaultTag is the default tag for updatecli OCI images.
	ociDefaultTag string = "latest"
)
