package cache

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/exitcode"
)

func TestNewSourceCache(t *testing.T) {
	// Arrange / Act
	c := NewSourceCache()

	// Assert
	require.NotNil(t, c)
	assert.Equal(t, 0, c.Len())
}

func TestSourceCache_GetSet(t *testing.T) {
	// Arrange
	c := NewSourceCache()
	key := "some-cache-key"
	want := SourceEntry{
		Information: "v1.2.3",
		Description: "latest stable release",
		Result:      "SUCCESS",
	}

	// Act
	c.Set(key, want)
	got, ok := c.Get(key)

	// Assert
	require.True(t, ok)
	assert.Equal(t, want.Information, got.Information)
	assert.Equal(t, want.Description, got.Description)
	assert.Equal(t, want.Result, got.Result)
}

func TestSourceCache_GetMiss(t *testing.T) {
	// Arrange
	c := NewSourceCache()

	// Act
	got, ok := c.Get("nonexistent-key")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, SourceEntry{}, got)
}

func TestSourceCache_EmptyKey(t *testing.T) {
	// Arrange
	c := NewSourceCache()
	entry := SourceEntry{
		Information: "some-value",
		Result:      "SUCCESS",
	}

	// Act: Set with empty key must be a no-op
	c.Set("", entry)

	// Assert: nothing was stored
	assert.Equal(t, 0, c.Len())

	// Act: Get with empty key must return false without panicking
	got, ok := c.Get("")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, SourceEntry{}, got)
}

func TestSourceCache_Overwrite(t *testing.T) {
	// Arrange
	c := NewSourceCache()
	key := "shared-key"
	first := SourceEntry{Information: "v1.0.0", Result: "SUCCESS"}
	second := SourceEntry{Information: "v2.0.0", Result: "SUCCESS"}

	// Act: write the same key twice
	c.Set(key, first)
	c.Set(key, second)
	got, ok := c.Get(key)

	// Assert: only the latest value is returned
	require.True(t, ok)
	assert.Equal(t, second.Information, got.Information)
	assert.Equal(t, 1, c.Len(), "overwriting an existing key must not grow the cache")
}

func TestSourceCache_Len(t *testing.T) {
	// Arrange
	c := NewSourceCache()
	entries := map[string]SourceEntry{
		"alpha": {Information: "1", Result: "SUCCESS"},
		"beta":  {Information: "2", Result: "SUCCESS"},
		"gamma": {Information: "3", Result: "FAILURE"},
	}

	// Act
	for k, v := range entries {
		c.Set(k, v)
	}

	// Assert
	assert.Equal(t, len(entries), c.Len())
}

// TestSourceCache_ConcurrentAccess verifies that concurrent reads and writes
// do not trigger the race detector.
func TestSourceCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	c := NewSourceCache()
	const workers = 20
	const opsPerWorker = 50

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for i := range workers {
		go func(i int) {
			defer wg.Done()
			for j := range opsPerWorker {
				key := "key-odd"
				if j%2 == 0 {
					key = "key-even"
				}
				c.Set(key, SourceEntry{
					Information: "value",
					Result:      "SUCCESS",
				})
				_ = i
			}
		}(i)
	}

	for range workers {
		go func() {
			defer wg.Done()
			for range opsPerWorker {
				c.Get("key-even")
				c.Len()
			}
		}()
	}

	wg.Wait()
}

// TestKey_EmptyKind verifies that Key returns an empty string when the
// ResourceConfig has no Kind. GetReportConfig cannot resolve an unknown plugin,
// so Key returns the empty-string sentinel that callers treat as a cache miss.
func TestKey_EmptyKind(t *testing.T) {
	// Arrange
	rc := resource.ResourceConfig{
		Kind: "",
		Name: "my-source",
	}

	// Act
	key := Key(rc)

	// Assert
	assert.Equal(t, "", key)
}

// shellSpec returns a minimal but fully-formed shell.Spec that resource.New()
// can instantiate without error, making it suitable as a Key() input.
func shellSpec(command string) shell.Spec {
	return shell.Spec{
		Command: command,
		ChangedIf: shell.SpecChangedIf{
			Kind: "exitcode",
			Spec: exitcode.Spec{Warning: 1, Success: 0, Failure: 2},
		},
	}
}

// TestKey_SameConfigProducesSameKey verifies the key is stable across two calls
// with identical config values.
func TestKey_SameConfigProducesSameKey(t *testing.T) {
	// Arrange: a fully-formed config that resource.New() can resolve.
	rc := resource.ResourceConfig{
		Kind: "shell",
		Name: "my-source",
		Spec: shellSpec("echo hello"),
	}

	// Act
	key1 := Key(rc)
	key2 := Key(rc)

	// Assert
	require.NotEmpty(t, key1)
	assert.Equal(t, key1, key2)
}

// TestKey_SameSpecDifferentNamesShareKey verifies that two configs with
// identical Kind+Spec but different Names produce the same key.
func TestKey_SameSpecDifferentNamesShareKey(t *testing.T) {
	spec := shellSpec("echo hello")
	rc1 := resource.ResourceConfig{Kind: "shell", Name: "name-a", Spec: spec}
	rc2 := resource.ResourceConfig{Kind: "shell", Name: "name-b", Spec: spec}

	key1 := Key(rc1)
	key2 := Key(rc2)

	require.NotEmpty(t, key1)
	assert.Equal(t, key1, key2)
}

// TestKey_DifferentSpecsProduceDifferentKeys verifies that two configs with
// different Spec values produce distinct keys.
func TestKey_DifferentSpecsProduceDifferentKeys(t *testing.T) {
	// Arrange
	rc1 := resource.ResourceConfig{Kind: "shell", Name: "source-a", Spec: shellSpec("echo a")}
	rc2 := resource.ResourceConfig{Kind: "shell", Name: "source-b", Spec: shellSpec("echo b")}

	// Act
	key1 := Key(rc1)
	key2 := Key(rc2)

	// Assert
	require.NotEmpty(t, key1)
	require.NotEmpty(t, key2)
	assert.NotEqual(t, key1, key2)
}
