package dockerhub

import (
	"errors"
	"os"
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
		expectedError:  errors.New("olblak/updatecli:donotexist not found on DockerHub"),
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
		expectedError:  errors.New("olblak/donotexist:latest not found on DockerHub"),
	},
	{
		docker: Docker{
			Image: "donotexist/donotexist",
			Tag:   "donotexist",
		},
		expectedDigest: "",
		expectedError:  errors.New("donotexist/donotexist:donotexist not found on DockerHub"),
	},
	{
		docker: Docker{
			Image: "jenkins/jenkins",
			Tag:   "doNotExist",
		},
		expectedDigest: "",
		expectedError:  errors.New("jenkins/jenkins:doNotExist not found on DockerHub"),
	},
	{
		docker: Docker{
			Image: "jenkins/jenkins",
			Tag:   "2.275",
			Token: os.Getenv("DOCKERHUB_TOKEN"),
		},
		expectedDigest: "e4630b9084110ad05b4b51f5131d62161881216d60433d1f2074d522c3dcd6dc",
	},
	{
		// Test private docker image with authentication
		docker: Docker{
			Image: "olblak/test",
			Tag:   "updatecli",
			Token: os.Getenv("DOCKERHUB_TOKEN"),
		},
		expectedDigest: "ce782db15ab5491c6c6178da8431b3db66988ccd11512034946a9667846952a6",
	},
	{
		// Test private docker image without authentication
		docker: Docker{
			Image: "olblak/test",
			Tag:   "updatecli",
		},
		expectedDigest: "",
		expectedError:  errors.New("olblak/test:updatecli not found on DockerHub"),
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
			t.Errorf("Docker Image %v:%v for arch %v, expect digest %v, got %v", d.docker.Image, d.docker.Tag, d.docker.Architecture, expected, got)
		}
	}
}
