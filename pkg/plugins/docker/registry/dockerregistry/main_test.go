package dockerregistry

import (
	"testing"

	"github.com/olblak/updateCli/pkg/core/helpers"

	"github.com/sirupsen/logrus"
)

type DataSet struct {
	docker         Docker
	expectedDigest string
}

var data = []DataSet{
	{
		docker: Docker{
			Image:    "azure-cli",
			Tag:      "2.0.27",
			Hostname: "mcr.microsoft.com",
			Client:   &helpers.DefaultHttpClient{},
		},
		expectedDigest: "d7c97a1951c336e4427450023409712a9993e8f1f8764be10e05e03d8c863279",
	},
	{
		docker: Docker{
			Image:    "azure-cli",
			Tag:      "donotexist",
			Hostname: "mcr.microsoft.com",
			Client:   &helpers.DefaultHttpClient{},
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image:    "dotnotexist",
			Tag:      "donotexist",
			Hostname: "mcr.microsoft.com",
			Client:   &helpers.DefaultHttpClient{},
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
			t.Errorf("Docker Image %v:%v expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}
