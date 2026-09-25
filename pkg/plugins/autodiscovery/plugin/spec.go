package plugin

/*
Spec defines the specification for a Wasm autodiscovery plugin, used by a crawler whose name ends with ".wasm".
The plugin runs in a sandbox and returns the manifests it generates.
*/
type Spec struct {
	// "spec" defines the plugin parameters.
	//
	// remark:
	//   * it is passed as is to the plugin, see the plugin documentation for the accepted keys.
	//
	Spec map[string]any `yaml:",omitempty"`
	// "allowedpaths" defines the paths the plugin can access from inside its sandbox.
	//
	// default:
	//   ```
	//   - ".:/mnt"
	//   ```
	//
	// remark:
	//   * a path is either a plain path or a "HOST_PATH:GUEST_PATH" mapping.
	//   * a relative host path is resolved from the scm directory when "scmid" is set, otherwise from the directory relative paths resolve from, by default the working directory.
	//   * by default, the plugin runs from "/mnt".
	//
	// example:
	//   ```
	//   allowedpaths:
	//     - .:/mnt
	//     - /var/lib/updatecli:/data
	//   ```
	//
	AllowedPaths *[]string `yaml:",omitempty"`
	// "allowhosts" defines the hosts the plugin can send HTTP requests to from inside its sandbox.
	//
	AllowHosts []string `yaml:",omitempty"`
	// "timeout" defines the maximum execution time of the plugin, in milliseconds.
	//
	// default:
	//   60000
	//
	// remark:
	//   * 0 disables the timeout.
	//
	Timeout *uint64 `yaml:",omitempty"`
}
