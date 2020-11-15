package dockerregistry

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
			Image:    "azure-cli",
			Tag:      "latest",
			Hostname: "mcr.microsoft.com",
		},
		expectedDigest: "bddcbadc711fd3c0a41c3101a0ba07ace4c1b124b6ee7b57be3ffe8142f140c9",
	},
	{
		docker: Docker{
			Image:    "azure-cli",
			Tag:      "donotexist",
			Hostname: "mcr.microsoft.com",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image:    "dotnotexist",
			Tag:      "donotexist",
			Hostname: "mcr.microsoft.com",
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
			t.Errorf("Docker Image %v:%v expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}
