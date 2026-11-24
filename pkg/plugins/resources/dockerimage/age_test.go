package dockerimage

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// testImage describes an image pushed to the test registry. An age of 0 leaves the date unset,
// a negative age sets it to the Unix epoch, as reproducible builds do.
type testImage struct {
	tag string
	// created is the age of the "created" field of the image configuration.
	created time.Duration
	// annotation is the age of the "org.opencontainers.image.created" manifest annotation.
	annotation time.Duration
	// platforms, when set, pushes an image index holding one image per platform instead.
	platforms map[string]testImage
	// platform, written "<os>/<architecture>", is set in the configuration of a single image.
	platform string
}

const day = 24 * time.Hour

// testRegistry is an in-memory registry counting the manifest requests it serves.
type testRegistry struct {
	host             string
	manifestRequests atomic.Int32
	// manifestsServed, when positive, answers later manifest requests with a bodyless 404,
	// as some registries do for a deleted tag.
	manifestsServed atomic.Int32
	// failures answers the manifest requests of a tag with a status code, set before any request.
	failures map[string]int
	// broken answers the manifest requests of a tag with a manifest that can't be parsed.
	broken map[string]bool
	// schema1 answers the manifest requests of a tag with a schema1 manifest.
	schema1 map[string]bool
	// missingChildren answers the manifest requests by digest with a 404, as for the images of
	// an index the registry no longer holds.
	missingChildren bool
	// requested holds the tags whose manifest was requested.
	requested sync.Map
}

func newTestRegistry(t *testing.T) *testRegistry {
	t.Helper()

	reg := &testRegistry{failures: map[string]int{}, broken: map[string]bool{}, schema1: map[string]bool{}}
	handler := registry.New(registry.Logger(stdlog.New(io.Discard, "", 0)))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/manifests/") {
			served := reg.manifestRequests.Add(1)
			if limit := reg.manifestsServed.Load(); limit > 0 && served > limit {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, tag, _ := strings.Cut(r.URL.Path, "/manifests/")
			reg.requested.Store(tag, true)
			if code, found := reg.failures[tag]; found {
				w.WriteHeader(code)
				return
			}
			if reg.missingChildren && strings.HasPrefix(tag, "sha256:") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if reg.schema1[tag] {
				w.Header().Set("Content-Type", string(types.DockerManifestSchema1))
				_, _ = w.Write([]byte("{}"))
				return
			}
			if reg.broken[tag] {
				w.Header().Set("Content-Type", string(types.OCIManifestSchema1))
				_, _ = w.Write([]byte("not a manifest"))
				return
			}
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	reg.host = u.Host

	return reg
}

func dateOf(ago time.Duration) time.Time {
	if ago < 0 {
		return time.Unix(0, 0)
	}
	return time.Now().Add(-ago)
}

func newImage(t *testing.T, spec testImage) v1.Image {
	t.Helper()

	img, err := random.Image(64, 1)
	require.NoError(t, err)

	if spec.created != 0 {
		img, err = mutate.CreatedAt(img, v1.Time{Time: dateOf(spec.created)})
		require.NoError(t, err)
	}

	if spec.platform != "" {
		config, err := img.ConfigFile()
		require.NoError(t, err)
		config = config.DeepCopy()
		config.OS, config.Architecture, _ = strings.Cut(spec.platform, "/")
		img, err = mutate.ConfigFile(img, config)
		require.NoError(t, err)
	}

	if spec.annotation != 0 {
		img = mutate.Annotations(img, map[string]string{
			createdAnnotation: dateOf(spec.annotation).Format(time.RFC3339),
		}).(v1.Image)
	}

	return img
}

func (reg *testRegistry) push(t *testing.T, images ...testImage) string {
	t.Helper()

	image := reg.host + "/app"

	for _, spec := range images {
		ref, err := name.ParseReference(image + ":" + spec.tag)
		require.NoError(t, err)

		if len(spec.platforms) == 0 {
			require.NoError(t, remote.Write(ref, newImage(t, spec)))
			continue
		}

		var index v1.ImageIndex = empty.Index
		for platform, platformSpec := range spec.platforms {
			os, arch, _ := strings.Cut(platform, "/")
			index = mutate.AppendManifests(index, mutate.IndexAddendum{
				Add: newImage(t, platformSpec),
				Descriptor: v1.Descriptor{
					Platform: &v1.Platform{OS: os, Architecture: arch},
				},
			})
		}

		if spec.annotation != 0 {
			index = mutate.Annotations(index, map[string]string{
				createdAnnotation: dateOf(spec.annotation).Format(time.RFC3339),
			}).(v1.ImageIndex)
		}

		require.NoError(t, remote.WriteIndex(ref, index))
	}

	reg.manifestRequests.Store(0)

	return image
}

var semverFilter = version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"}

func TestSourceAge(t *testing.T) {
	tests := []struct {
		name              string
		images            []testImage
		age               age.Spec
		architecture      string
		wantTag           string
		wantResult        string
		wantInDescription string
	}{
		{
			name: "the newest tag is too recent, the previous one is selected",
			images: []testImage{
				{tag: "v1.0.0", created: 30 * day},
				{tag: "v1.1.0", created: 10 * day},
				{tag: "v1.2.0", created: 1 * day},
			},
			age:        age.Spec{Minimum: "7d"},
			wantTag:    "v1.1.0",
			wantResult: result.SUCCESS,
		},
		{
			name: "every tag is too recent",
			images: []testImage{
				{tag: "v1.1.0", created: 2 * day},
				{tag: "v1.2.0", created: 1 * day},
			},
			age:               age.Spec{Minimum: "7d"},
			wantResult:        result.SKIPPED,
			wantInDescription: "matches the age filter yet",
		},
		{
			name: "every tag is too old",
			images: []testImage{
				{tag: "v1.1.0", created: 30 * day},
				{tag: "v1.2.0", created: 10 * day},
			},
			age:               age.Spec{Maximum: "5d"},
			wantResult:        result.SKIPPED,
			wantInDescription: "matches the age filter yet",
		},
		{
			name: "the annotation wins over the configuration",
			images: []testImage{
				{tag: "v1.1.0", created: 10 * day},
				{tag: "v1.2.0", created: 30 * day, annotation: 1 * day},
			},
			age:        age.Spec{Minimum: "7d"},
			wantTag:    "v1.1.0",
			wantResult: result.SUCCESS,
		},
		{
			name: "a tag created at the Unix epoch is undated",
			images: []testImage{
				{tag: "v1.1.0", created: -1},
				{tag: "v1.2.0", created: -1},
			},
			age:               age.Spec{Minimum: "7d"},
			wantResult:        result.SKIPPED,
			wantInDescription: "reports a creation date",
		},
		{
			name: "an undated tag is passed over",
			images: []testImage{
				{tag: "v1.1.0", created: 10 * day},
				{tag: "v1.2.0", created: -1},
			},
			age:        age.Spec{Minimum: "7d"},
			wantTag:    "v1.1.0",
			wantResult: result.SUCCESS,
		},
		{
			name: "the image of the requested platform is dated",
			images: []testImage{
				{tag: "v1.1.0", created: 30 * day},
				{tag: "v1.2.0", platforms: map[string]testImage{
					"linux/amd64": {created: 1 * day},
					"linux/arm64": {created: 10 * day},
				}},
			},
			age:          age.Spec{Minimum: "7d"},
			architecture: "linux/arm64",
			wantTag:      "v1.2.0",
			wantResult:   result.SUCCESS,
		},
		{
			name: "the image of another platform is too recent",
			images: []testImage{
				{tag: "v1.1.0", created: 30 * day},
				{tag: "v1.2.0", platforms: map[string]testImage{
					"linux/amd64": {created: 1 * day},
					"linux/arm64": {created: 10 * day},
				}},
			},
			age:          age.Spec{Minimum: "7d"},
			architecture: "linux/amd64",
			wantTag:      "v1.1.0",
			wantResult:   result.SUCCESS,
		},
		{
			name: "the index annotation wins over the platform image",
			images: []testImage{
				{tag: "v1.1.0", created: 30 * day},
				{tag: "v1.2.0", annotation: 1 * day, platforms: map[string]testImage{
					"linux/arm64": {created: 10 * day},
				}},
			},
			age:          age.Spec{Minimum: "7d"},
			architecture: "linux/arm64",
			wantTag:      "v1.1.0",
			wantResult:   result.SUCCESS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newTestRegistry(t)
			image := reg.push(t, tt.images...)

			spec := Spec{Image: image, VersionFilter: semverFilter, Age: tt.age}
			if tt.architecture != "" {
				spec.Architecture = tt.architecture
			}

			di, err := New(spec)
			require.NoError(t, err)

			got := result.Source{}
			err = di.Source(context.Background(), pathresolver.Resolver{}, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.wantResult, got.Result)
			assert.Equal(t, tt.wantTag, got.Information)
			assert.Contains(t, got.Description, tt.wantInDescription)
		})
	}
}

func TestSourceAgeBoundsTheChecks(t *testing.T) {
	reg := newTestRegistry(t)

	var images []testImage
	for i := range age.MaxChecks + 5 {
		images = append(images, testImage{tag: "v1." + string(rune('a'+i)) + ".0", created: 1 * day})
	}
	image := reg.push(t, images...)

	di, err := New(Spec{
		Image:         image,
		VersionFilter: version.Filter{Kind: version.REGEXVERSIONKIND, Pattern: `^v1\.\w\.0$`},
		Age:           age.Spec{Minimum: "7d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, result.SKIPPED, got.Result)
	assert.Equal(t, int32(age.MaxChecks), reg.manifestRequests.Load())
}

func TestSourceAgeRegistryErrors(t *testing.T) {
	tests := []struct {
		name              string
		failures          map[string]int
		wantTag           string
		wantResult        string
		wantInDescription string
		wantErr           string
	}{
		{
			name:              "a failure after a tag too recent skips rather than selecting a lower tag",
			failures:          map[string]int{"v1.1.0": http.StatusInternalServerError},
			wantResult:        result.SKIPPED,
			wantInDescription: "registry failed to answer",
		},
		{
			name:       "a deleted tag is passed over",
			failures:   map[string]int{"v1.1.0": http.StatusNotFound},
			wantTag:    "v1.0.0",
			wantResult: result.SUCCESS,
		},
		{
			name:              "rate limiting skips",
			failures:          map[string]int{"v1.2.0": http.StatusTooManyRequests},
			wantResult:        result.SKIPPED,
			wantInDescription: "registry failed to answer",
		},
		{
			name:              "a failure after a deleted tag skips, as the registry answered",
			failures:          map[string]int{"v1.2.0": http.StatusNotFound, "v1.1.0": http.StatusBadGateway},
			wantResult:        result.SKIPPED,
			wantInDescription: "registry failed to answer",
		},
		{
			name:     "a registry failing for the first tag fails",
			failures: map[string]int{"v1.2.0": http.StatusUnauthorized},
			wantErr:  "v1.2.0",
		},
		{
			name: "no tag found fails",
			failures: map[string]int{
				"v1.2.0": http.StatusNotFound,
				"v1.1.0": http.StatusNotFound,
				"v1.0.0": http.StatusNotFound,
			},
			wantErr: "none of the 3 tags checked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newTestRegistry(t)
			image := reg.push(t,
				testImage{tag: "v1.0.0", created: 30 * day},
				testImage{tag: "v1.1.0", created: 10 * day},
				testImage{tag: "v1.2.0", created: 1 * day},
			)
			reg.failures = tt.failures

			di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
			require.NoError(t, err)

			got := result.Source{}
			err = di.Source(context.Background(), pathresolver.Resolver{}, &got)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantResult, got.Result)
			assert.Equal(t, tt.wantTag, got.Information)
			assert.Contains(t, got.Description, tt.wantInDescription)
		})
	}
}

func TestSourceAgeTruncatedManifest(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", created: 30 * day},
		testImage{tag: "v1.1.0", created: 10 * day},
		testImage{tag: "v1.2.0", created: 1 * day},
	)
	reg.broken["v1.1.0"] = true

	di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	// A manifest that can't be parsed may be a truncated answer, so v1.0.0 isn't selected instead
	assert.Equal(t, result.SKIPPED, got.Result)
	assert.Contains(t, got.Description, "registry failed to answer")
}

func TestSourceAgeSchema1PassedOver(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", created: 30 * day},
		testImage{tag: "v1.1.0", created: 10 * day},
		testImage{tag: "v1.2.0", created: 1 * day},
	)
	reg.schema1["v1.1.0"] = true

	di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, result.SUCCESS, got.Result)
	assert.Equal(t, "v1.0.0", got.Information)
}

func TestSourceAgeCancelled(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t, testImage{tag: "v1.0.0", created: 30 * day})

	di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = di.Condition(ctx, "v1.0.0", nil, pathresolver.Resolver{})

	assert.ErrorIs(t, err, context.Canceled)
}

func TestSourceAgeEveryImageBroken(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", created: 30 * day},
		testImage{tag: "v1.1.0", created: 30 * day},
	)
	reg.schema1["v1.0.0"] = true
	reg.schema1["v1.1.0"] = true

	di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, result.SKIPPED, got.Result)
	assert.Contains(t, got.Description, "is supported")
	assert.NotContains(t, got.Description, "remove it")
}

func TestSourceAgeIndexMissingItsImage(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", platforms: map[string]testImage{"linux/amd64": {created: 30 * day}}},
		testImage{tag: "v1.1.0", platforms: map[string]testImage{"linux/amd64": {created: 30 * day}}},
	)
	reg.missingChildren = true

	di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	got := result.Source{}
	err = di.Source(context.Background(), pathresolver.Resolver{}, &got)

	// Both tags are passed over as missing their image, rather than failing as a registry error
	assert.ErrorContains(t, err, "none of the 2 tags checked")
	assert.NotContains(t, err.Error(), "dating the tag")
}

func TestConfigSatisfies(t *testing.T) {
	tests := []struct {
		name   string
		config v1.ConfigFile
		arch   string
		want   bool
	}{
		{name: "same architecture", config: v1.ConfigFile{OS: "linux", Architecture: "amd64"}, arch: "amd64", want: true},
		{name: "other architecture", config: v1.ConfigFile{OS: "linux", Architecture: "arm64"}, arch: "amd64"},
		{name: "no architecture in the configuration", config: v1.ConfigFile{}, arch: "amd64", want: true},
		{name: "no variant in the configuration", config: v1.ConfigFile{OS: "linux", Architecture: "arm"}, arch: "linux/arm/v7", want: true},
		{name: "other variant", config: v1.ConfigFile{OS: "linux", Architecture: "arm", Variant: "v6"}, arch: "linux/arm/v7"},
		{name: "operating system not requested", config: v1.ConfigFile{OS: "windows", Architecture: "amd64"}, arch: "amd64", want: true},
		{name: "other operating system requested", config: v1.ConfigFile{OS: "windows", Architecture: "amd64"}, arch: "linux/amd64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, configSatisfies(&tt.config, tt.arch))
		})
	}
}

func TestSourceAgeChecksTheFilterBeforeListing(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t, testImage{tag: "v1.0.0", created: 30 * day})

	var listed atomic.Bool
	di, err := New(Spec{Image: image, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)
	di.options = append(di.options, remote.WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		listed.Store(true)
		return http.DefaultTransport.RoundTrip(r)
	})))

	got := result.Source{}
	err = di.Source(context.Background(), pathresolver.Resolver{}, &got)

	assert.ErrorIs(t, err, errAgeNeedsOrderedFilter)
	assert.False(t, listed.Load(), "the registry mustn't be queried")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSourceAgeDatesThePreferredTagAlone(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", created: 30 * day},
		testImage{tag: "v1.1.0", created: 30 * day},
		testImage{tag: "v1.2.0", created: 30 * day},
	)

	di, err := New(Spec{Image: image, VersionFilter: semverFilter, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, "v1.2.0", got.Information)
	assert.Equal(t, int32(1), reg.manifestRequests.Load())
}

func TestIsRegistryError(t *testing.T) {
	assert.True(t, isRegistryError(&transport.Error{StatusCode: http.StatusBadGateway}))
	assert.True(t, isRegistryError(fmt.Errorf("wrapped: %w", &url.Error{Op: "Get", URL: "https://example.com", Err: io.EOF})))
	assert.True(t, isRegistryError(io.ErrUnexpectedEOF), "a dropped connection may not happen on the next run")
	assert.False(t, isRegistryError(fmt.Errorf("%w: schema1", errUnsupportedImage)))
}

func TestSourceAgeKeepsParsedVersion(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", created: 30 * day},
		testImage{tag: "v1.1.0", created: 1 * day},
	)

	di, err := New(Spec{
		Image:         image,
		VersionFilter: version.Filter{Kind: version.REGEXSEMVERVERSIONKIND, Regex: `^v(\d+\.\d+\.\d+)$`},
		Age:           age.Spec{Minimum: "7d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, "v1.0.0", got.Information)
	assert.Equal(t, "1.0.0", di.foundVersion.ParsedVersion)
}

func TestSourceAgeFetchesATagOnce(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t, testImage{tag: "v1.0.0", annotation: 30 * day, platforms: map[string]testImage{
		"linux/amd64": {created: 30 * day},
		"linux/arm64": {created: 30 * day},
	}})

	di, err := New(Spec{
		Image:         image,
		Architectures: []string{"linux/amd64", "linux/arm64"},
		VersionFilter: semverFilter,
		Age:           age.Spec{Minimum: "7d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, "v1.0.0", got.Information)
	// The index annotation dates every platform, so only the index is requested
	assert.Equal(t, int32(1), reg.manifestRequests.Load())
}

func TestPlatformManifest(t *testing.T) {
	image := func(os, arch string) v1.Descriptor {
		return v1.Descriptor{MediaType: types.OCIManifestSchema1, Platform: &v1.Platform{OS: os, Architecture: arch}}
	}

	tests := []struct {
		name      string
		manifests []v1.Descriptor
		arch      string
		want      string
		wantFound bool
	}{
		{
			name:      "linux/amd64 is preferred without a platform",
			manifests: []v1.Descriptor{image("unknown", "unknown"), image("linux", "arm64"), image("linux", "amd64")},
			want:      "amd64",
			wantFound: true,
		},
		{
			name:      "the first image without linux/amd64, attestations aside",
			manifests: []v1.Descriptor{image("unknown", "unknown"), image("linux", "arm64"), image("linux", "s390x")},
			want:      "arm64",
			wantFound: true,
		},
		{
			name:      "the requested platform",
			manifests: []v1.Descriptor{image("linux", "amd64"), image("linux", "arm64")},
			arch:      "arm64",
			want:      "arm64",
			wantFound: true,
		},
		{
			name:      "a missing platform",
			manifests: []v1.Descriptor{image("linux", "amd64")},
			arch:      "linux/arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := platformManifest(tt.manifests, tt.arch)

			require.Equal(t, tt.wantFound, found)
			if found {
				assert.Equal(t, tt.want, got.Platform.Architecture)
			}
		})
	}
}

func TestSourceAgeStopsWhenRateLimited(t *testing.T) {
	reg := newTestRegistry(t)

	var images []testImage
	for i := range 8 {
		tag := "v1." + string(rune('a'+i)) + ".0"
		images = append(images, testImage{tag: tag, created: 1 * day})
	}
	image := reg.push(t, images...)
	for _, img := range images {
		reg.failures[img.tag] = http.StatusTooManyRequests
	}

	di, err := New(Spec{
		Image:         image,
		VersionFilter: version.Filter{Kind: version.REGEXVERSIONKIND, Pattern: `^v1\.\w\.0$`},
		Age:           age.Spec{Minimum: "7d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	// The registry client retries rate limited requests, so tags are counted rather than requests
	requested := 0
	reg.requested.Range(func(_, _ any) bool {
		requested++
		return true
	})

	assert.Equal(t, result.SKIPPED, got.Result)
	// The preferred tag is dated alone, and rate limiting stops the search
	assert.Equal(t, 1, requested)
}

func TestSourceAgeEveryArchitecture(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", platforms: map[string]testImage{
			"linux/amd64": {created: 40 * day},
			"linux/arm64": {created: 40 * day},
		}},
		testImage{tag: "v1.1.0", platforms: map[string]testImage{
			"linux/amd64": {created: 31 * day},
			"linux/arm64": {created: 2 * day},
		}},
		testImage{tag: "v1.2.0", platforms: map[string]testImage{
			"linux/amd64": {created: 40 * day},
		}},
	)

	di, err := New(Spec{
		Image:         image,
		Architectures: []string{"linux/amd64", "linux/arm64"},
		VersionFilter: semverFilter,
		Age:           age.Spec{Minimum: "30d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	// v1.2.0 has no arm64 image and v1.1.0 is too recent on arm64
	assert.Equal(t, result.SUCCESS, got.Result)
	assert.Equal(t, "v1.0.0", got.Information)
}

func TestSourceAgeNeedsOrderedFilter(t *testing.T) {
	reg := newTestRegistry(t)
	image := reg.push(t, testImage{tag: "v1.0.0", created: 10 * day})

	di, err := New(Spec{Image: image, Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	got := result.Source{}
	err = di.Source(context.Background(), pathresolver.Resolver{}, &got)

	assert.ErrorIs(t, err, errAgeNeedsOrderedFilter)
}

func TestSourceAgeTimeFilterWithoutVersion(t *testing.T) {
	reg := newTestRegistry(t)
	// The time filter parses the lowercase month but looks the tag up as "2026-Feb-01".
	image := reg.push(t, testImage{tag: "2026-feb-01", created: 10 * day})

	di, err := New(Spec{
		Image:         image,
		VersionFilter: version.Filter{Kind: version.TIMEVERSIONKIND, Pattern: "2006-Jan-02"},
		Age:           age.Spec{Minimum: "7d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	err = di.Source(context.Background(), pathresolver.Resolver{}, &got)

	assert.Error(t, err)
}

func TestConditionAge(t *testing.T) {
	tests := []struct {
		name          string
		tag           string
		architectures []string
		want          bool
		wantMessage   string
	}{
		{name: "old enough", tag: "v1.1.0", want: true},
		{name: "too recent", tag: "v1.2.0", wantMessage: "outside the age filter"},
		{name: "undated", tag: "v1.0.0", wantMessage: "no creation date"},
		{name: "old enough for every platform", tag: "v2.0.0", architectures: []string{"linux/arm64"}, want: true},
		{name: "too recent for one platform", tag: "v2.0.0", architectures: []string{"linux/arm64", "linux/amd64"}, wantMessage: "outside the age filter"},
		{name: "missing platform", tag: "v2.0.0", architectures: []string{"linux/s390x"}, wantMessage: "not found"},
		{name: "missing tag", tag: "v9.0.0", wantMessage: "not found"},
		{name: "index without linux/amd64 and no architecture", tag: "v3.0.0", want: true},
		{name: "single image of the requested platform", tag: "v4.0.0", architectures: []string{"linux/arm64"}, want: true},
		{name: "single image of another platform", tag: "v4.0.0", architectures: []string{"linux/amd64"}, wantMessage: "not found"},
	}

	reg := newTestRegistry(t)
	image := reg.push(t,
		testImage{tag: "v1.0.0", created: -1},
		testImage{tag: "v1.1.0", created: 10 * day},
		testImage{tag: "v1.2.0", created: 1 * day},
		testImage{tag: "v2.0.0", platforms: map[string]testImage{
			"linux/amd64": {created: 1 * day},
			"linux/arm64": {created: 10 * day},
		}},
		testImage{tag: "v3.0.0", platforms: map[string]testImage{
			"linux/arm64": {created: 10 * day},
		}},
		testImage{tag: "v4.0.0", created: 10 * day, platform: "linux/arm64"},
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di, err := New(Spec{
				Image:         image,
				Tag:           tt.tag,
				Architectures: tt.architectures,
				Age:           age.Spec{Minimum: "7d"},
			})
			require.NoError(t, err)

			got, message, err := di.Condition(context.Background(), "", nil, pathresolver.Resolver{})
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
			assert.Contains(t, message, tt.wantMessage)
		})
	}
}

func TestNewInvalidAge(t *testing.T) {
	_, err := New(Spec{Image: "ghcr.io/updatecli/updatecli", Age: age.Spec{Minimum: "7"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong age spec")
}
