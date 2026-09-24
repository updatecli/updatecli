package temurin

/*
"temurin" defines the specification for retrieving Eclipse Temurin releases from the Adoptium API.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "releaseline" defines the line of Temurin release to retrieve.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   lts
	//
	// remark:
	//   * accepted values are "lts" and "feature".
	//   * only used when neither "featureversion" nor "specificversion" is set.
	//
	// example:
	//   * releaseline: feature
	//
	ReleaseLine string `yaml:",omitempty"`
	// "releasetype" defines the type of Temurin release to retrieve.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   ga
	//
	// remark:
	//   * accepted values are "ga" (general availability) and "ea" (early access, such as nightly builds).
	//
	// example:
	//   * releasetype: ea
	//
	ReleaseType string `yaml:",omitempty"`
	// "featureversion" defines the major Java version used to filter the Temurin releases.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   the most recent release of the "releaseline".
	//
	// remark:
	//   * accepted values are integers such as 8, 11, 17 or 21.
	//   * "featureversion" and "specificversion" are mutually exclusive.
	//
	// example:
	//   * featureversion: 21
	//
	FeatureVersion int `yaml:",omitempty"`
	// "result" defines the type of value returned from the retrieved Temurin release.
	//
	// compatible:
	//   * source
	//
	// default:
	//   version
	//
	// remark:
	//   * accepted values are:
	//     * "version": the version name, which is the Temurin scm release name.
	//     * "installer_url": the HTTP URL of the binary release or installer.
	//     * "checksum_url": the HTTP URL of the checksum file.
	//     * "signature_url": the HTTP URL of the signature file.
	//
	// example:
	//   * result: installer_url
	//
	Result string `yaml:",omitempty"`
	// "architecture" defines the CPU architecture used to filter the Temurin releases.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   x64
	//
	// remark:
	//   * accepted values are defined by the Temurin API, see https://api.adoptium.net/q/swagger-ui/#/Types
	//   * common values are:
	//     * "x64" (Intel/AMD 64 bits)
	//     * "x86" (Intel/AMD 32 bits)
	//     * "ppc64" (PowerPC 64 bits)
	//     * "ppc64le" (PowerPC little endian 64 bits)
	//     * "s390x" (IBM Z)
	//     * "aarch64" (ARM 64 bits)
	//     * "arm" (ARM 32 bits)
	//     * "sparcv9" (Sparc 64 bits)
	//     * "riscv64" (RISC-V 64 bits)
	//   * "architecture" and "platforms" are mutually exclusive.
	//
	// example:
	//   * architecture: aarch64
	//
	Architecture string `yaml:",omitempty"`
	// "imagetype" defines the type of artifact used to filter the Temurin releases.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   jdk
	//
	// remark:
	//   * accepted values are "jdk", "jre", "testimage", "debugimage", "staticlibs", "source" and "sbom".
	//
	// example:
	//   * imagetype: jre
	//
	ImageType string `yaml:",omitempty"`
	// "operatingsystem" defines the operating system used to filter the Temurin releases.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   linux
	//
	// remark:
	//   * accepted values are defined by the Temurin API, see https://api.adoptium.net/q/swagger-ui/#/Types
	//   * common values are "linux", "windows", "mac", "solaris", "aix" and "alpine-linux".
	//   * "operatingsystem" and "platforms" are mutually exclusive.
	//
	// example:
	//   * operatingsystem: windows
	//
	OperatingSystem string `yaml:",omitempty"`
	// "specificversion" defines the exact Temurin version used to filter the Temurin releases.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   empty
	//
	// remark:
	//   * accepted values are a semantic version, a JDK version or a Temurin release name.
	//   * "featureversion" and "specificversion" are mutually exclusive.
	//   * in a condition, the source output overrides it unless "disablesourceinput" is set to true.
	//
	// example:
	//   * specificversion: 17.0.2+8
	//
	SpecificVersion string `yaml:",omitempty"`
	// "project" defines the project used to filter the Temurin releases.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   jdk
	//
	// remark:
	//   * accepted values are "jdk", "valhalla", "metropolis", "jfr" and "shenandoah".
	//
	// example:
	//   * project: valhalla
	//
	Project string `yaml:",omitempty"`
	// "platforms" defines a list of platforms used to filter the Temurin releases.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   empty, so no filtering per platform.
	//
	// remark:
	//   * each platform combines an operating system and a CPU architecture separated by a slash ("/").
	//   * accepted operating systems and architectures are defined by the Temurin API, see https://api.adoptium.net/q/swagger-ui/#/Types
	//   * "platforms" is mutually exclusive with "architecture" and "operatingsystem".
	//
	// example:
	//   * platforms:
	//     - linux/x64
	//     - linux/aarch64
	//     - linux/s390x
	//     - alpine-linux/x64
	//     - windows/x64
	//
	Platforms []string `yaml:",omitempty"`
}
