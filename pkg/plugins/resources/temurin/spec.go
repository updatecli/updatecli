package temurin

type Spec struct {
	// ReleaseLine specifies the line of Temurin release to retrieve.
	//
	// default: "lts"
	//
	// Allowed values:
	// * "lts"
	// * "feature"
	ReleaseLine string `yaml:",omitempty"`
	// ReleaseType specifies the type of Temurin release to retrieve.
	//
	// default: "ga"
	//
	// Allowed values:
	// * "ga" (General Availability)
	// * "ea" (Early Availability, e.g. nightly builds)
	ReleaseType string `yaml:",omitempty"`
	// FeatureVersion specifies the Major Java version to filter the Temurin release to retrieve.
	//
	// default: undefined
	//
	// Allowed values: integer number (8, 11, 17, 21, etc.)
	FeatureVersion int `yaml:",omitempty"`
	// Result specifies the type of value returned by the retrieved Temurin release.
	//
	// default: "version"
	//
	// Allowed values:
	// * "version" (Version Name, e.g. the Temurin SCM release name)
	// * "installer_url" (HTTP URL to the binary release/installer)
	// * "checksum_url" (HTTP URL to the checksum file)
	// * "signature_url" (HTTP URL to the signature file)
	Result string `yaml:",omitempty"`
	// Architecture specifies the CPU architecture to filter the Temurin release to retrieve.
	//
	// default: "x64"
	//
	// Allowed values:
	// * "x64" (Intel/AMD 64 Bits)
	// * "x86" (Intel/AMD 32 Bits)
	// * "ppc64" (PowerPC 64 Bits)
	// * "ppc64le" (PowerPC Little Endian 64 Bits)
	// * "s390x" (IBM Z)
	// * "aarch64" (ARM 64 Bits)
	// * "arm" (ARM 32 Bits)
	// * "sparcv9" (Sparc 64 Bits)
	// * "riscv64" (RiscV 64 Bits)
	Architecture string `yaml:",omitempty"`
	// ImageType specifies the type of artifact to filter the Temurin release to retrieve.
	//
	// default: "jdk"
	//
	// Allowed values:
	// * "jdk"
	// * "jre"
	// * "testimage"
	// * "debugimage"
	// * "staticlibs"
	// * "source
	// * "sbom"
	ImageType string `yaml:",omitempty"`
	// OperatingSystem specifies the Operating System to filter the Temurin release to retrieve.
	//
	// default: "linux"
	//
	// Allowed values:
	// * "linux"
	// * "windows"
	// * "mac"
	// * "solaris"
	// * "aix"
	// * "alpine-linux"
	OperatingSystem string `yaml:",omitempty"`
	// SpecificVersion specifies the exact Temurin version to filter the Temurin release to retrieve.
	// Ignores FeatureVersion when used.
	//
	// default: undefined
	//
	// Allowed values: string (can be a semantic version, a JDK version or a temurin release name)
	SpecificVersion string `yaml:",omitempty"`
	// Project specifies the project to filter the Temurin release to retrieve.
	//
	// default: "jdk"
	//
	// Allowed values:
	// * "jdk" (default)
	// * "valhalla"
	// * "metropolis"
	// * "jfr"
	// * "shenandoah"
	Project string `yaml:",omitempty"`
}
