package githubrelease

// Changelog returns the content (body) of the Github Release
func (gr GitHubRelease) Changelog() string {
	changelog, _ := gr.ghHandler.Changelog(gr.foundVersion)
	return changelog
}
