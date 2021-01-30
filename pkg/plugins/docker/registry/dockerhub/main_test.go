package dockerhub

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

type DataSet struct {
	docker         Docker
	expectedDigest string
}

var data = []DataSet{
	{
		docker: Docker{
			Image:        "olblak/updatecli",
			Tag:          "v0.0.16",
			Architecture: "amd64",
		},
		expectedDigest: "3c615fb45d190c8dfcdc8cb6b020aa27b86610755694d3ef072495d368ef81e5",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "v0.0.16",
		},
		expectedDigest: "3c615fb45d190c8dfcdc8cb6b020aa27b86610755694d3ef072495d368ef81e5",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "donotexist",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "library/nginx",
			Tag:   "1.12.1",
		},
		expectedDigest: "0f5baf09c628c0f44c1d53be8293f95ee80cd542f2ea37c48a667d535614b12a",
	},
	{
		docker: Docker{
			Image:        "library/nginx",
			Tag:          "1.12.1",
			Architecture: "arm64",
		},
		expectedDigest: "5ac3c8e77726c5f6eb2268f6a6914fca53c91f73506e28c6a86fe34cd8e3f468",
	},
	{
		docker: Docker{
			Image: "olblak/donotexist",
			Tag:   "latest",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "donotexist/donotexist",
			Tag:   "donotexist",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "jenkins/jenkins",
			Tag:   "doNotExist",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "jenkins/jenkins",
			Tag:   "2.275",
			Token: os.Getenv("DOCKERHUB_TOKEN"),
		},
		expectedDigest: "e4630b9084110ad05b4b51f5131d62161881216d60433d1f2074d522c3dcd6dc",
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
			t.Errorf("Docker Image %v:%v for arch %v, expect digest %v, got %v", d.docker.Image, d.docker.Tag, d.docker.Architecture, expected, got)
		}
	}
}
