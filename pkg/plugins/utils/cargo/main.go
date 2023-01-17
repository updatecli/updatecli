package cargo

type InlineKeyChain struct {
	// [A][S][C] Token specifies the cargo registry token to use for authentication.
	Token string `yaml:",omitempty"`
}
