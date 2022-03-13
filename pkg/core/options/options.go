package options

// Options hold target parameters
type Pipeline struct {
	Commit bool
	Push   bool
	Clean  bool
	DryRun bool
}

// Option contains configuration options such as filepath located on disk,etc.
type Config struct {
	// ManifestFile contains the updatecli manifest full file path
	ManifestFile string
	// ValuesFiles contains the list of updatecli values full file path
	ValuesFiles []string
	// SecretsFiles contains the list of updatecli sops secrets full file path
	SecretsFiles []string
	// DisableTemplating specify if needs to be done
	DisableTemplating bool
}
