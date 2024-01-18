package config

var (
	// Define indentation used to encode yaml data
	YAMLSetIdent int = 4

	/*
		GolangTemplatingDiff is used to enable or disable the diff feature.
		Showing the diff may leak sensitive information like credentials.
	*/
	GolangTemplatingDiff bool
)
