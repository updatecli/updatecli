package version

// TODO: Remove this struct once https://github.com/updatecli/updatecli/issues/803 is fixed.
// Version defines a version from a filter that holds both the original found version and the parsed version (depending on the kind of filter: semantic, text, etc.)
// Keeping the original found versions is useful when checking for metadata around the version, such as the changelog
type Version struct {
	ParsedVersion   string
	OriginalVersion string
}

// TODO: Change the receiver of this function to Filter once https://github.com/updatecli/updatecli/issues/803 is fixed.
func (v Version) GetVersion() string {
	return v.OriginalVersion
}
