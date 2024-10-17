package keywords

type Logic interface {
	IsLineMatching(originalLine, matcher string) bool
	ReplaceLine(source, originalLine, matcher string) string
	GetValue(originalLine, matcher string) (string, error)
	GetTokens(originalLine string) (Tokens, error)
}

type Tokens interface {
}
