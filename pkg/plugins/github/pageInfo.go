package github

// PageInfo is used for Graphql queries to iterate over pagination
type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	EndCursor       string
	StartCursor     string
}
