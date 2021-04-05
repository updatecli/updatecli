package quay

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
			Image:  "jetstack/cert-manager-controller",
			Tag:    "v1.0.0",
			Client: &helpers.DefaultHttpClient{},
		},
		expectedDigest: "8eda7cd9fe3e72fd23c9646fd6e4fba5407113872462268aa37ae3660eda9992",
	},
	{
		docker: Docker{
			Image:  "coreos/prometheus-operator",
			Tag:    "v0.39.0-arm64",
			Client: &helpers.DefaultHttpClient{},
		},
		expectedDigest: "3142406cb96f300355462312607db40078828b32e9c7a904c3e461687383b96",
	},
	{
		docker: Docker{
			Image:  "coreos/prometheus-operator",
			Tag:    "v0.39.0-amd64",
			Client: &helpers.DefaultHttpClient{},
		},
		expectedDigest: "775ed3360c67ae11a2521a404b0964f65245bbd75f498118cd4058e09c8fcb91",
	},
	{
		docker: Docker{
			Image:  "jetstack/cert-manager-controller",
			Tag:    "donotexist",
			Client: &helpers.DefaultHttpClient{},
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image:  "jetstack/donotexist",
			Tag:    "donotexist",
			Client: &helpers.DefaultHttpClient{},
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image:  "donotexist/donotexist",
			Tag:    "donotexist",
			Client: &helpers.DefaultHttpClient{},
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
