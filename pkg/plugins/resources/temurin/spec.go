package temurin

type Spec struct {
	// ReleaseLine specifies the type of Temurin Release to retrieve: 'lts' (default), feature.
	ReleaseLine string `yaml:",omitempty"`
	// ReleaseType specifies the type of Temurin release: 'ga' (default), 'ea' (nightly releases).
	ReleaseType string `yaml:",omitempty"`
	// FeatureVersion specifies the Java Feature version (major) to filter results.
	FeatureVersion int `yaml:",omitempty"`
	// Result specifies the type of result returned by the resource: 'version' (default), installer_url, checksum/
	Result string `yaml:",omitempty"`
	// Architecture specifies the CPU architecture as per Temurin conventions: 'x64' (default), 'x86', 'x32', 'ppc64', 'ppc64le', 's390x', 'aarch64', 'arm', 'sparcv9', 'riscv64'.
	Architecture string `yaml:",omitempty"`
	// ImageType specifies the type of artifact as per Temurin conventions: 'jdk' (default), 'jre', 'testimage', 'debugimage', 'staticlibs', 'sources', 'sbom'
	ImageType string `yaml:",omitempty"`
	// OperatingSystem specifies the type of Operating System as per Temurin conventions: 'linux' (default), 'windows', 'mac', 'solaris', 'aix', 'alpine-linux'
	OperatingSystem string `yaml:",omitempty"`
	// SpecificVersion specifies the exact Temurin version instead of latest. Ignores FeatureVersion when used.
	SpecificVersion string `yaml:",omitempty"`
}
