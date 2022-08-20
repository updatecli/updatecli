package githubrelease

var (
	DeprecatedSemverVersionMessage string = `Deprecated behavior detected
|++++
| If the Github Release you are interacting with, doesn't prepend the retrieved version with the character "v" such as "v1.0.0",
| then feel free to add the parameter 'originalVersion: true' to hide this message such as in the following example.
| Otherwise please read this message carefully.
| 
| Before:
| ----
| < sources:
| <     example:
| <         name: Get latest Updatecli release
| <         kind: githubrelease
| <         spec:
| <             owner: updatecli
| <             repository: updatecli
| <             token: '{{ requiredEnv "GITHUB_TOKEN" }}'
| <             username: '{{ requiredEnv "GITHUB_ACTOR" }}'
| <             versionfilter:
| <                 kind: semver
| ----
|
| After:
| ----
| > sources:
| >     example:
| >         name: Get latest Updatecli release
| >         kind: githubrelease
| >         spec:
| >             owner: updatecli
| >             repository: updatecli
| >             token: '{{ requiredEnv "GITHUB_TOKEN" }}'
| >             username: '{{ requiredEnv "GITHUB_ACTOR" }}'
| >             keeporiginalversion: true
| >             versionfilter:
| >                 kind: semver
| >             transformers:
| >                 - trimprefix: v
| ----
| 
| In Updatecli, we consider that information retrieved from a "Source" shouldn't be altered.
| Instead, it should be the role of a "Transformer" rule to modify it.
|
| The Github Release version retrieve using the semantic version rule will stop dropping the "v" if one is specified in the GitHub release tag such as v1.0.0.
| This bug has been around for so long that we saw a significant amount of Updatecli manifest written in a way that we can't automatically fix by running "updatecli manifest upgrade"
| 
| As we take seriously backward compatibility for Updatecli manifest, we introduce the parameter "keeporiginalversion" to keep the current deprecated behavior until all manifests are updated.
| 
| Please adapt your manifest to this new behavior and then set "keeporiginalversion" to true to validate that you are manifest is compatible with the new behavior.
| More information on https://github.com/updatecli/updatecli/issues/803
|++++
`
)
