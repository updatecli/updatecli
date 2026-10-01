package githubsearch

import (
	"context"
	"reflect"
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient answers the repository search with repositories and lists the branches main
// and dev for each of them, counting the queries it receives.
type mockClient struct {
	repositories []string
	queries      int
}

func (m *mockClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	m.queries++
	query := reflect.ValueOf(q).Elem()

	if search := query.FieldByName("Search"); search.IsValid() {
		edges := search.FieldByName("Edges")
		edges.Set(reflect.MakeSlice(edges.Type(), len(m.repositories), len(m.repositories)))
		for i, repository := range m.repositories {
			edges.Index(i).FieldByName("Node").FieldByName("Repository").FieldByName("NameWithOwner").SetString(repository)
		}
		return nil
	}

	nodes := query.FieldByName("Repository").FieldByName("Refs").FieldByName("Nodes")
	nodes.Set(reflect.MakeSlice(nodes.Type(), 2, 2))
	nodes.Index(0).FieldByName("Name").SetString("main")
	nodes.Index(1).FieldByName("Name").SetString("dev")
	return nil
}

func (m *mockClient) Mutate(ctx context.Context, mutation interface{}, input githubv4.Input, variables map[string]interface{}) error {
	return nil
}

func TestScmsGeneratorDiscoversOnce(t *testing.T) {
	client := &mockClient{repositories: []string{"updatecli/updatecli", "updatecli/website"}}

	newSearch := func(search string) GitHubSearch {
		return GitHubSearch{
			spec:   Spec{Search: search, URL: "https://github.example.com/discover-once"},
			branch: "^main$",
			search: search,
			client: client,
		}
	}

	expectedRepositories := []string{"updatecli", "website"}

	// One search query, then one branch query per repository
	for range 3 {
		specs, err := newSearch("org:updatecli").ScmsGenerator(context.Background())
		require.NoError(t, err)

		got := []string{}
		for _, spec := range specs {
			assert.Equal(t, "main", spec.Branch)
			assert.Equal(t, "updatecli", spec.Owner)
			got = append(got, spec.Repository)
		}
		assert.Equal(t, expectedRepositories, got)
	}
	assert.Equal(t, 3, client.queries, "the same discovery should only query GitHub once")

	_, err := newSearch("org:updatecli archived:false").ScmsGenerator(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 4, client.queries, "another search should only rerun the search, not the branch listings")
}
