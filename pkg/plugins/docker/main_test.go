package docker

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

type DataSet struct {
	docker             Docker
	expectedHostname   string
	expectedImage      string
	expectedDigest     string
	expectedError      error
	expectedCondition  bool
	requireCredentials bool
}

var data = []DataSet{
	// Test Dockerhub
	{
		docker: Docker{
			Image:    "olblak/test",
			Tag:      "updatecli",
			Username: os.Getenv("DOCKERHUB_USERNAME"),
			Password: os.Getenv("DOCKERHUB_PASSWORD"),
		},
		expectedCondition:  true,
		expectedDigest:     "ce782db15ab5491c6c6178da8431b3db66988ccd11512034946a9667846952a6",
		expectedHostname:   "hub.docker.com",
		expectedImage:      "olblak/test",
		requireCredentials: true,
	},
	{
		docker: Docker{
			Image: "olblak/test",
			Tag:   "updatecli",
		},
		expectedCondition: false,
		expectedDigest:    "",
		expectedHostname:  "hub.docker.com",
		expectedImage:     "olblak/test",
		expectedError:     errors.New("olblak/test:updatecli not found on DockerHub"),
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "v0.0.16",
		},
		expectedCondition: true,
		expectedDigest:    "3c615fb45d190c8dfcdc8cb6b020aa27b86610755694d3ef072495d368ef81e5",
		expectedHostname:  "hub.docker.com",
		expectedImage:     "olblak/updatecli",
	},
	{
		docker: Docker{
			Image: "olblak/updatecli",
			Tag:   "donotexist",
		},
		expectedCondition: false,
		expectedDigest:    "",
		expectedHostname:  "hub.docker.com",
		expectedImage:     "olblak/updatecli",
		expectedError:     errors.New("olblak/updatecli:donotexist not found on DockerHub"),
	},
	{
		docker: Docker{
			Image: "nginx",
			Tag:   "1.12.1",
		},
		expectedCondition: true,
		expectedHostname:  "hub.docker.com",
		expectedImage:     "library/nginx",
		expectedDigest:    "0f5baf09c628c0f44c1d53be8293f95ee80cd542f2ea37c48a667d535614b12a",
	},
	// Test OCI registry
	{
		docker: Docker{
			Image: "mcr.microsoft.com/azure-cli",
			Tag:   "2.0.27",
		},
		expectedCondition: true,
		expectedHostname:  "mcr.microsoft.com",
		expectedImage:     "azure-cli",
		expectedDigest:    "d7c97a1951c336e4427450023409712a9993e8f1f8764be10e05e03d8c863279",
	},
	// Test ghcr registry
	{
		docker: Docker{
			Image:    "ghcr.io/olblak/updatecli",
			Tag:      "v0.0.22",
			Username: os.Getenv("GITHUB_ACTOR"),
			Password: os.Getenv("GITHUB_TOKEN"),
		},
		expectedCondition: true,
		expectedHostname:  "ghcr.io",
		expectedImage:     "olblak/updatecli",
		expectedDigest:    "f237aed76d3d00538d44448e8161df00d6c044f8823cc8eb9aeccc8413f5a029",
	},
	{
		// Test that not providing a tag, fallback to latest
		docker: Docker{
			Image:    "ghcr.io/olblak/updatecli",
			Username: os.Getenv("GITHUB_ACTOR"),
			Password: os.Getenv("GITHUB_TOKEN"),
		},
		expectedCondition: true,
		expectedHostname:  "ghcr.io",
		expectedImage:     "olblak/updatecli",
		expectedDigest:    "975ce2b7e362c2689bb11cdcee7ef84aefd00bf2f4123771722a639635c231d4",
	},
	{
		docker: Docker{
			Image:    "ghcr.io/olblak/updatecli",
			Tag:      "donotexist",
			Username: os.Getenv("GITHUB_ACTOR"),
			Password: os.Getenv("GITHUB_TOKEN"),
		},
		expectedCondition: false,
		expectedHostname:  "ghcr.io",
		expectedError:     errors.New("ghcr.io/olblak/updatecli:donotexist - repository name not known to registry"),
		expectedImage:     "olblak/updatecli",
		expectedDigest:    "",
	},
	{
		docker: Docker{
			Image:    "ghcr.io/donotexist/donotexist",
			Tag:      "donotexist",
			Username: os.Getenv("GITHUB_ACTOR"),
			Password: os.Getenv("GITHUB_TOKEN"),
		},
		expectedCondition: false,
		expectedHostname:  "ghcr.io",
		expectedError:     errors.New("ghcr.io/donotexist/donotexist:donotexist - repository name not known to registry"),
		expectedImage:     "donotexist/donotexist",
		expectedDigest:    "",
	},

	// Test quay
	{
		docker: Docker{
			Image: "quay.io/jetstack/cert-manager-controller",
			Tag:   "v1.0.0",
		},
		expectedCondition: true,
		expectedHostname:  "quay.io",
		expectedImage:     "jetstack/cert-manager-controller",
		expectedDigest:    "8eda7cd9fe3e72fd23c9646fd6e4fba5407113872462268aa37ae3660eda9992",
	},
	{
		docker: Docker{
			Image: "quay.io/jetstack/cert-manager-controller",
			Tag:   "v1.0.0",
			Token: "WrongToken",
		},
		expectedCondition: true,
		expectedHostname:  "quay.io",
		expectedImage:     "jetstack/cert-manager-controller",
		expectedDigest:    "8eda7cd9fe3e72fd23c9646fd6e4fba5407113872462268aa37ae3660eda9992",
	},
	{
		docker: Docker{
			Image: "quay.io/jetstack/cert-manager-controller",
			Tag:   "donotexist",
		},
		expectedCondition: false,
		expectedHostname:  "quay.io",
		expectedImage:     "jetstack/cert-manager-controller",
		expectedDigest:    "",
		expectedError:     errors.New("tag doesn't exist for quay.io/jetstack/cert-manager-controller:donotexist"),
	},
	{
		docker: Docker{
			Image: "quay.io/donotexist/donotexist",
			Tag:   "donotexist",
		},
		expectedCondition: false,
		expectedHostname:  "quay.io",
		expectedImage:     "donotexist/donotexist",
		expectedDigest:    "",
		expectedError:     errors.New("quay.io/donotexist/donotexist:donotexist - doesn't exist on quay.io"),
	},
}

func TestParseImage(t *testing.T) {
	for _, d := range data {
		if testing.Short() && d.requireCredentials {
			t.Skip("Skipping test in short mode when it requires specific credentials")
			continue
		}
		hostnameGot, imageGot, err := parseImage(d.docker.Image)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}

		if hostnameGot != d.expectedHostname {
			t.Errorf("Wrong hostname found! expected %v, got %v", d.expectedHostname, hostnameGot)
		}

		if imageGot != d.expectedImage {
			t.Errorf("Wrong image found! expected %v, got %v", d.expectedImage, imageGot)
		}

	}
}

func TestParameters(t *testing.T) {
	d := &Docker{
		Image: "nginx",
	}

	errs := d.Validate()

	if len(errs) > 0 {
		for _, err := range errs {
			t.Log(err)
		}
		t.Errorf("Minimum valid configuration provided! Expect %v, got %v", true, false)
	}

	// Test if we correctly return the default docker hub url if not defined
	expected := "latest"
	got := d.Tag
	if got != expected {
		t.Errorf("Tag is not configured! expected %v, got %v", expected, got)
	}
}

func TestCondition(t *testing.T) {
	// Test if existing image tag return true

	for _, d := range data {
		// Short mode also skip integration test that require credentials
		if testing.Short() && d.requireCredentials {
			t.Skip("Skipping test in short mode when it requires specific credentials")
			continue
		}

		got, err := d.docker.Condition("")

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

		if got != d.expectedCondition {
			t.Errorf("Is %v:%v published?\nExpected:\t\t%v\nGot:\t\t\t%v",
				d.docker.Image, d.docker.Tag,
				d.expectedCondition, got)
		}
	}
}

func TestSource(t *testing.T) {
	// Test if existing return the correct digest
	for _, d := range data {
		// Short mode also skip integration test that require credentials
		if testing.Short() && d.requireCredentials {
			t.Skip("Skipping test in short mode when it requires specific credentials")
			continue
		}

		got, err := d.docker.Source("")

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
		if strings.Compare(got, expected) != 0 {
			t.Errorf("Testing Docker Image %v:%v:\nExpected:\t\t%v\nGot:\t\t\t%v",
				d.docker.Image, d.docker.Tag,
				expected, got)
		}
	}
}
