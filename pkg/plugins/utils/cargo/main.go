package cargo

type InlineKeyChain struct {
	// [A][S][C] Token specifies the cargo registry token to use for authentication.
	Token string `yaml:",omitempty"`
	// [A][S][C] HeaderFormat specifies the cargo registry header format to use for authentication (defaults to `Bearer`).
	HeaderFormat string `yaml:"headerFormat,omitempty"`
}

type Registry struct {
	// [A][S][C] Auth specifies the cargo registry auth to use for authentication.
	Auth InlineKeyChain `yaml:",omitempty"`
	// [A][S][C] URL specifies the cargo registry URL to use for authentication.
	URL string `yaml:",omitempty"`
	// [A][S][C] RootDir specifies the cargo registry root directory to use as FS index.
	RootDir string `yaml:",omitempty"`
	// [A] SCMID specifies the cargo registry scmId to use as FS index.
	SCMID string `yaml:",omitempty"`
}
