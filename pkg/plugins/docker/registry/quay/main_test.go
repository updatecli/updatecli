package quay

import (
	"errors"
	"strings"
	"testing"
)

type DataSet struct {
	docker         Docker
	expectedDigest string
	expectedError  error
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
			Image: "coreos/prometheus-operator",
			Tag:   "v0.39.0-arm64",
		},
		expectedDigest: "3142406cb96f300355462312607db40078828b32e9c7a904c3e461687383b96",
	},
	{
		docker: Docker{
			Image: "coreos/prometheus-operator",
			Tag:   "v0.39.0-amd64",
		},
		expectedDigest: "775ed3360c67ae11a2521a404b0964f65245bbd75f498118cd4058e09c8fcb91",
	},
	{
		docker: Docker{
			Image: "jetstack/cert-manager-controller",
			Tag:   "donotexist",
		},
		expectedDigest: "",
		expectedError:  errors.New("tag doesn't exist for quay.io/jetstack/cert-manager-controller:donotexist"),
	},
	{
		docker: Docker{
			Image: "jetstack/donotexist",
			Tag:   "donotexist",
		},
		expectedDigest: "",
		expectedError:  errors.New("quay.io/jetstack/donotexist:donotexist - doesn't exist on quay.io"),
	},
	{
		docker: Docker{
			Image: "donotexist/donotexist",
			Tag:   "donotexist",
		},
		expectedDigest: "",
		expectedError:  errors.New("quay.io/donotexist/donotexist:donotexist - doesn't exist on quay.io"),
	},
}

func TestDigest(t *testing.T) {
	// Test if existing return the correct digest
	for _, d := range data {
		got, err := d.docker.Digest()

		if err != nil && d.expectedError != nil {
			if strings.Compare(err.Error(), d.expectedError.Error()) != 0 {

				t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q\n",
					d.expectedError.Error(), err.Error())
			}
		} else if err != nil && d.expectedError == nil {
			t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q\n",
				"nil", err.Error())

		} else if err == nil && d.expectedError != nil {
			t.Errorf("Unexpected error:\nExpected:\t\t%q\nGot:\t\t\t%q\n",
				d.expectedError.Error(), "nil")
		}

		expected := d.expectedDigest
		if got != expected {
			t.Errorf("Docker Image %v:%v expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}
