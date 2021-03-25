package ghcr

import (
	"github.com/olblak/updateCli/pkg/core/helpers"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDigest(t *testing.T) {
	var (
		testCases = []struct {
			docker         Docker
			expectedDigest string
		}{
			{
				docker: Docker{
					Image:  "olblak/updatecli",
					Tag:    "v0.0.25",
					Token:  os.Getenv("GITHUB_TOKEN"),
					Client: &helpers.FakeHttpClient{
						Requests: map[string]helpers.FakeResponse{
							"https://ghcr.io/v2/olblak/updatecli/manifests/v0.0.25": {
								StatusCode: 200,
								Body: GetContents("test_data/0.0.25.json"),
								Headers: map[string][]string{
									"Docker-Content-Digest":{"sha256:786e49e87808a9808625cfca69b86e8e4e6a26d7f6199499f927633ea906676f"},
								},
							},
						},
					},
				},
				expectedDigest: "786e49e87808a9808625cfca69b86e8e4e6a26d7f6199499f927633ea906676f",
			},
			{
				docker: Docker{
					Image: "olblak/updatecli",
					Tag:   "v0.0.22",
					Token:  os.Getenv("GITHUB_TOKEN"),
					Client: &helpers.FakeHttpClient{
						Requests: map[string]helpers.FakeResponse{
							"https://ghcr.io/v2/olblak/updatecli/manifests/v0.0.22": {
								StatusCode: 200,
								Body: GetContents("test_data/0.0.22.json"),
								Headers: map[string][]string{
									"Docker-Content-Digest":{"sha256:f237aed76d3d00538d44448e8161df00d6c044f8823cc8eb9aeccc8413f5a029"},
								},
							},
						},
					},
				},
				expectedDigest: "f237aed76d3d00538d44448e8161df00d6c044f8823cc8eb9aeccc8413f5a029",
			},
			{
				docker: Docker{
					Image:  "olblak/updatecli",
					Tag:    "v0.0.24",
					Token:  os.Getenv("GITHUB_TOKEN"),
					Client: &helpers.FakeHttpClient{
						Requests: map[string]helpers.FakeResponse{
							"https://ghcr.io/v2/olblak/updatecli/manifests/v0.0.24": {
								StatusCode: 200,
								Body: GetContents("test_data/0.0.24.json"),
								Headers: map[string][]string{
									"Docker-Content-Digest":{"sha256:a0dfa59bddbaa538f40e2ef8eb7d87cc7591b3e2d725a1bec9135ed304f88053"},
								},
							},
						},
					},
				},
				expectedDigest: "a0dfa59bddbaa538f40e2ef8eb7d87cc7591b3e2d725a1bec9135ed304f88053",
			},
			{
				docker: Docker{
					Image:  "olblak/updatecli",
					Tag:    "donotexist",
					Token:  os.Getenv("GITHUB_TOKEN"),
					Client: &helpers.FakeHttpClient{
						Requests: map[string]helpers.FakeResponse{
							"https://ghcr.io/v2/olblak/updatecli/manifests/donotexist": {
								StatusCode: 404,
								Body: GetContents("test_data/unknown.json"),
							},
						},
					},
				},
				expectedDigest: "",
			},
			{
				docker: Docker{
					Image:  "donotexist/donotexist",
					Tag:    "donotexist",
					Token:  os.Getenv("GITHUB_TOKEN"),
					Client: &helpers.FakeHttpClient{
						Requests: map[string]helpers.FakeResponse{
							"https://ghcr.io/v2/donotexist/donotexist/manifests/donotexist": {
								StatusCode: 404,
								Body: GetContents("test_data/unknown.json"),
							},
						},
					},
				},
				expectedDigest: "",
			},
		}
	)

	// Test if existing return the correct digest
	for _, d := range testCases {
		t.Run(d.docker.Image, func(t *testing.T) {
			got, err := d.docker.Digest()

			if err != nil {
				logrus.Errorf("err - %s", err)
			}
			expected := d.expectedDigest

			if got != expected {
				t.Errorf("Docker Image %v:%v, expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
			}
		})
	}
}

func GetContents(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}
