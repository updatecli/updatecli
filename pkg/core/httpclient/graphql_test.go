package httpclient

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQLTransport_MarksQueriesOnly(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "query with variables", body: `{"query":"query($owner:String!){repository(owner:$owner){id}}"}`, want: true},
		{name: "query without variables", body: `{"query":"{viewer{login}}"}`, want: true},
		{name: "mutation", body: `{"query":"mutation($input:CreatePullRequestInput!){createPullRequest(input:$input){clientMutationId}}"}`, want: false},
		{name: "mutation after a comment", body: `{"query":"# create PR\nmutation{createPullRequest(input:{}){clientMutationId}}"}`, want: false},
		{name: "mutation after a fragment", body: `{"query":"fragment F on PullRequest{id} mutation{createPullRequest(input:{}){pullRequest{...F}}}"}`, want: false},
		{name: "not JSON", body: `not json`, want: false},
		{name: "empty query", body: `{}`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange — a transport that records what it receives.
			var got *http.Request
			var gotBody string
			transport := &graphQLTransport{transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				got = req
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				gotBody = string(b)
				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
			})}

			req, err := http.NewRequest(http.MethodPost, "http://example.invalid/graphql", strings.NewReader(tt.body))
			require.NoError(t, err)

			// Act
			_, err = transport.RoundTrip(req)
			require.NoError(t, err)

			// Assert — only queries are marked, and the body still reaches the server.
			assert.Equal(t, tt.want, isIdempotent(got))
			assert.Equal(t, tt.body, gotBody)
		})
	}
}
