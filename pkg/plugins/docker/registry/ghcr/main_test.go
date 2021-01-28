package ghcr

import (
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

type DataSet struct {
	docker         Docker
	expectedDigest string
}

var data = []DataSet{
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "v0.0.25",
			Token: os.Getenv("GITHUB_TOKEN"),
		},
		expectedDigest: "786e49e87808a9808625cfca69b86e8e4e6a26d7f6199499f927633ea906676f",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "v0.0.22",
			Token: os.Getenv("GITHUB_TOKEN"),
		},
		expectedDigest: "f237aed76d3d00538d44448e8161df00d6c044f8823cc8eb9aeccc8413f5a029",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "v0.0.24",
			Token: os.Getenv("GITHUB_TOKEN"),
		},
		expectedDigest: "a0dfa59bddbaa538f40e2ef8eb7d87cc7591b3e2d725a1bec9135ed304f88053",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "donotexist",
			Token: os.Getenv("GITHUB_TOKEN"),
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "donotexist/donotexist",
			Tag:   "donotexist",
			Token: os.Getenv("GITHUB_TOKEN"),
		},
		expectedDigest: "",
	},
}

func TestDigest(t *testing.T) {
	// Test if existing return the correct digest
	for _, d := range data {
		got, err := d.docker.Digest()

		if err != nil {
			logrus.Errorf("err - %s", err)
		}
		expected := d.expectedDigest

		if got != expected {
			t.Errorf("Docker Image %v:%v, expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}
