package quay

import (
	"fmt"
	"testing"
)

type DataSet struct {
	docker         Docker
	digest         string
	expectedDigest string
}

var data = []DataSet{
	{
		docker: Docker{
			Image: "jetstack/cert-manager-controller",
			Tag:   "v1.0.0",
		},
		expectedDigest: "8eda7cd9fe3e72fd23c9646fd6e4fba5407113872462268aa37ae3660eda9992",
	},
	{
		docker: Docker{
			Image: "jetstack/cert-manager-controller",
			Tag:   "donotexist",
		},
		expectedDigest: "",
	},
	{
		docker: Docker{
			Image: "jetstack/donotexist",
			Tag:   "donotexist",
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
