package dockerimage

import (
	"context"
	"io"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
}

const day = 24 * time.Hour

// testRegistry is an in-memory registry counting the manifest requests it serves.
type testRegistry struct {
	host             string
	manifestRequests atomic.Int32
}

func newTestRegistry(t *testing.T) *testRegistry {
	t.Helper()

	reg := &testRegistry{}
	handler := registry.New(registry.Logger(stdlog.New(io.Discard, "", 0)))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/manifests/") {
			reg.manifestRequests.Add(1)
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

	if spec.annotation != 0 {
		img = mutate.Annotations(img, map[string]string{
			createdAnnotation: dateOf(spec.annotation).Format(time.RFC3339),
		}).(v1.Image)
	}

	return img
}

func (reg *testRegistry) push(t *testing.T, repository string, images ...testImage) string {
	t.Helper()

	image := reg.host + "/" + repository

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
			image := reg.push(t, "app", tt.images...)

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
	for i := range maxAgeChecks + 5 {
		images = append(images, testImage{tag: "v1." + string(rune('a'+i)) + ".0", created: 1 * day})
	}
	image := reg.push(t, "app", images...)

	di, err := New(Spec{
		Image:         image,
		VersionFilter: version.Filter{Kind: version.REGEXVERSIONKIND, Pattern: `^v1\.\w\.0$`},
		Age:           age.Spec{Minimum: "7d"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, di.Source(context.Background(), pathresolver.Resolver{}, &got))

	assert.Equal(t, result.SKIPPED, got.Result)
	assert.Equal(t, int32(maxAgeChecks), reg.manifestRequests.Load())
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
	}

	reg := newTestRegistry(t)
	image := reg.push(t, "app",
		testImage{tag: "v1.0.0", created: -1},
		testImage{tag: "v1.1.0", created: 10 * day},
		testImage{tag: "v1.2.0", created: 1 * day},
		testImage{tag: "v2.0.0", platforms: map[string]testImage{
			"linux/amd64": {created: 1 * day},
			"linux/arm64": {created: 10 * day},
		}},
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
