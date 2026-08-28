package config

import "fmt"

// RelativePathBase names the directory the relative paths of a manifest resolve against.
type RelativePathBase string

const (
	// RelativePathBaseUndefined is the zero value, meaning the manifest did not pick a
	// base and the one configured globally applies.
	RelativePathBaseUndefined RelativePathBase = ""
	// RelativePathBaseWorkingDirectory resolves relative paths against the directory
	// updatecli was started from. This is the historical behavior and the default.
	RelativePathBaseWorkingDirectory RelativePathBase = "workingdirectory"
	// RelativePathBaseManifest resolves relative paths against the directory holding the
	// manifest that declared them, which makes a manifest relocatable.
	RelativePathBaseManifest RelativePathBase = "manifest"
)

// ManifestOptions groups the manifest level settings that change how Updatecli behaves,
// as opposed to the keys describing what the pipeline is made of.
//
// It is named ManifestOptions rather than Options because config.Option already names the
// options a caller passes to New, which are a different thing: how to read a manifest off
// disk, not how the pipeline it describes behaves. The two barely overlap — "partialfiles"
// is meaningless inside a manifest — so they stay separate types.
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
	// 	* it only affects resources that are not attached to an "scm". A resource with an
	// 	  "scmid" always resolves its paths against the scm working directory.
	// 	* it overrides the "--relative-paths" command line flag.
	RelativePaths RelativePathBase `yaml:",omitempty" jsonschema:"enum=workingdirectory,enum=manifest"`
}

// Merge fills the settings the manifest left undefined with the ones configured globally.
//
// Each setting decides for itself: "relativepaths" lets the manifest override the command
// line, while a setting that exists to prevent something (such as a future "dryrun") would
// have to OR the two so the safer value always wins. That is why this is a method rather
// than a plain struct assignment.
func (o *ManifestOptions) Merge(defaults ManifestOptions) {
	if o.RelativePaths == RelativePathBaseUndefined {
		o.RelativePaths = defaults.RelativePaths
	}
}

// Validate reports a setting Updatecli does not understand, rather than silently falling
// back to the default.
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
