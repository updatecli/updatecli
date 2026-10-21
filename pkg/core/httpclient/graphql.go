package httpclient

import (
	"encoding/json"
	"net/http"
	"strings"
)

// graphQLTransport marks GraphQL queries as idempotent before passing them on.
// GraphQL sends queries and mutations alike as POST, so the retry client would
// otherwise never retry a query after a transport error, a 502, a 503 or a 504.
type graphQLTransport struct {
	transport http.RoundTripper
}

func (t *graphQLTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if isGraphQLQuery(req) {
		req = Idempotent(req)
	}
	return t.transport.RoundTrip(req)
}

// isGraphQLQuery reports whether req carries a GraphQL query, as opposed to a
// mutation. It reads a copy of the body, so req is left untouched. Only a
// document that starts with a query operation counts, as githubv4 writes it:
// anything else, such as a leading comment or fragment, is treated as a
// mutation, which is never retried. A request that names an operation is
// treated as a mutation too, as the named operation may be a mutation later
// in a document that starts with a query.
func isGraphQLQuery(req *http.Request) bool {
	if req.GetBody == nil {
		return false
	}

	body, err := req.GetBody()
	if err != nil {
		return false
	}
	defer body.Close()

	var payload struct {
		Query         string `json:"query"`
		OperationName string `json:"operationName"`
	}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return false
	}

	if payload.OperationName != "" {
		return false
	}

	query := strings.TrimSpace(payload.Query)
	return strings.HasPrefix(query, "query") || strings.HasPrefix(query, "{")
}

// NewUncachedGraphQLClient returns an HTTP client for a GraphQL API that
// bypasses the HTTP cache, like NewUncachedRetryClient, and retries queries
// as it retries a GET while sending mutations only once.
func NewUncachedGraphQLClient() *http.Client {
	return &http.Client{Transport: &graphQLTransport{transport: UncachedTransport()}}
}
