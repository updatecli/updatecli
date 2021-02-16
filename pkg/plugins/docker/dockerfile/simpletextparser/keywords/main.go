package keywords

type Logic interface {
	IsLineMatching(originalLine, matcher string) bool
	ReplaceLine(source, originalLine, matcher string) string
}
