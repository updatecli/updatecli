package config

import "fmt"

// RelativePathBase names the directory the relative paths of a manifest resolve against.
type RelativePathBase string

const (
	// RelativePathBaseUndefined is the zero value, meaning the manifest did not pick a
	// base and the one configured globally applies.
	RelativePathBaseUndefined RelativePathBase = ""
	// RelativePathBaseWorkingDirectory resolves relative paths against the directory
	// updatecli was started from. This is the default.
	RelativePathBaseWorkingDirectory RelativePathBase = "workingdirectory"
	// RelativePathBaseManifest resolves relative paths against the directory holding the
	// manifest that declared them, which makes a manifest relocatable.
	RelativePathBaseManifest RelativePathBase = "manifest"
)

// ManifestOptions groups the manifest level settings that change how Updatecli behaves,
// as opposed to the keys describing what the pipeline is made of.
//
// It is not called Options because config.Option already names the options that control
// how a manifest is read from disk.
type ManifestOptions struct {
	// "relativepaths" defines what the relative paths of this manifest resolve against.
	//
	// accepted values:
	// 	* "workingdirectory": relative to the directory updatecli was started from. Default.
	// 	* "manifest": relative to the directory holding this manifest.
	//
	// example:
	// ---
	// options:
	//   relativepaths: manifest
	// sources:
	//   version:
	//     kind: json
	//     spec:
	//       # resolved next to this manifest rather than next to the shell that ran updatecli
	//       file: package.json
	//       key: .version
	// ---
	//
	// remark:
	// 	* a resource with an "scmid" resolves its file paths against the scm checkout,
	// 	  whatever this setting says.
	// 	* with "manifest", a relative scm "directory" also resolves from the manifest
	// 	  directory, and so does the "path" of the gittag, gitbranch and gitcommit
	// 	  resources, even when they have an "scmid".
	// 	* with "manifest", autodiscovery without an scm searches the manifest directory,
	// 	  and the "local" scm is guessed from the git repository holding the manifest. A
	// 	  manifest outside a git repository cannot use "scmid: local".
	// 	* it overrides the "--relative-paths" command line flag.
	RelativePaths RelativePathBase `yaml:",omitempty" jsonschema:"enum=workingdirectory,enum=manifest"`
}

// Merge fills the settings the manifest left undefined with the ones configured globally.
//
// A value set in the manifest wins over the command line.
func (o *ManifestOptions) Merge(defaults ManifestOptions) {
	if o.RelativePaths == RelativePathBaseUndefined {
		o.RelativePaths = defaults.RelativePaths
	}
}

// Validate returns an error for a setting Updatecli does not understand, so a typo never
// silently falls back to the default.
func (o ManifestOptions) Validate() error {
	switch o.RelativePaths {
	case RelativePathBaseUndefined, RelativePathBaseWorkingDirectory, RelativePathBaseManifest:
		//
	default:
		return fmt.Errorf("%q is not a valid value for %q, expecting one of %q or %q",
			o.RelativePaths,
			"relativepaths",
			RelativePathBaseWorkingDirectory,
			RelativePathBaseManifest)
	}

	return nil
}
