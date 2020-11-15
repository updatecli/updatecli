package ghcr

import (
	"fmt"
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
			Tag:   "v0.0.22",
			Token: "xxx",
		},
		expectedDigest: "fd0a342a6df8b4ecb10b38c16a222bc3a964be1ab34547dbf116910b2184f4b9",
	},
	{
		docker: Docker{
			Image:        "olblak/updatecli",
			Tag:          "v0.0.22",
			Token:        "xxx",
			Architecture: "amd64",
		},
		expectedDigest: "fd0a342a6df8b4ecb10b38c16a222bc3a964be1ab34547dbf116910b2184f4b9",
	},
	{
		docker: Docker{
			Image:        "olblak/updatecli",
			Tag:          "v0.0.22",
			Token:        "xxx",
			Architecture: "arm64",
		},
		expectedDigest: "8c2d98213a7851b8dd8f3851f7453ab0ea5c9f92ca495882ef193466c3a92c21",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "donotexist",
			Token: "xxx",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "donotexist/donotexist",
			Tag:   "donotexist",
			Token: "xxx",
		},
		expectedDigest: "",
	},
}

func TestDigest(t *testing.T) {
	// Test if existing return the correct digest
	for _, d := range data {
		got, err := d.docker.Digest()

		if err != nil {
			fmt.Println(err)
		}
		expected := d.expectedDigest

		if got != expected {
			t.Errorf("Docker Image %v:%v for architecture '%s', expect digest %v, got %v", d.docker.Image, d.docker.Tag, d.docker.Architecture, expected, got)
		}
	}
}
