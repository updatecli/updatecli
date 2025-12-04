package plugin

type Spec struct {
	// Spec contains the plugin parameters.
	// cfr the plugin documentation.
	Spec map[string]any `yaml:",omitempty"`
	// AllowPaths is list of path to be accessed from inside the plugin sandbox,
	// a path can be either a plain path or a map from HOST_PATH:GUEST_PATH
	//
	// Example:
	//   - .:/mnt
	//   - /var/lib/updatecli:/data
	//
	// Default: [".:/mnt"]
	//
	// Remark:
	//   * Relative path are considered relative to the Updatecli working directory.
	//     If a scm root directory is set, relative paths are considered relative to the scm root directory.
	//   * By default, the plugin run from "/mnt"
	AllowedPaths *[]string `yaml:",omitempty"`
	// AllowedHosts hold a list of allowed hosts for HTTP requests from the plugin sandbox
	AllowHosts []string `yaml:",omitempty"`
	// Timeout defines a maximum execution time for the plugin in seconds
	//
	// Default: 300 seconds
	Timeout *uint64 `yaml:",omitempty"`
}
