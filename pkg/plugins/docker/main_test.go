package docker

import (
	"fmt"
	"os"
	"testing"
)

type DataSet struct {
	docker            Docker
	digest            string
	expectedHostname  string
	expectedImage     string
	expectedDigest    string
	expectedCondition bool
}

var data = []DataSet{
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
	{
		docker: Docker{
			Image: "ghcr.io/olblak/updatecli",
			Tag:   "v0.0.22",
			Token: os.Getenv("GITHUB_TOKEN"),
		},
		expectedCondition: true,
		expectedHostname:  "ghcr.io",
		expectedImage:     "olblak/updatecli",
		expectedDigest:    "f237aed76d3d00538d44448e8161df00d6c044f8823cc8eb9aeccc8413f5a029",
	},
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
			Tag:   "donotexist",
		},
		expectedCondition: false,
		expectedHostname:  "quay.io",
		expectedImage:     "jetstack/cert-manager-controller",
		expectedDigest:    "",
	},
}

func TestParseImage(t *testing.T) {
	for _, d := range data {
		hostnameGot, imageGot, err := parseImage(d.docker.Image)
		if err != nil {
			fmt.Println(err)
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

	ok, _ := d.Check()

	// Test if current setting is valid
	if ok != true {
		t.Errorf("Minimum valid configuration provided! Expect %v, got %v", true, ok)
	}

	// Test if we correctly return the default docker hub url if not defined
	expected := "latest"
	got := d.Tag
	if got != expected {
		t.Errorf("Tag is not configured! expected %v, got %v", expected, got)
	}
}

func TestCheck(t *testing.T) {

	// Test if image is not defined
	d := &Docker{}

	expected := false
	got, _ := d.Check()
	if got != false {
		t.Errorf("Image is not configured! expected %v, got %v", expected, got)
	}
}

func TestCondition(t *testing.T) {
	// Test if existing image tag return true

	for _, d := range data {
		got, _ := d.docker.Condition("")
		expected := d.expectedCondition
		if got != expected && expected {
			t.Errorf("%v:%v is published! expected %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		} else if got != expected && !expected {
			t.Errorf("%v:%v is not published! expected %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}

func TestSource(t *testing.T) {
	// Test if existing return the correct digest
	for _, d := range data {
		got, _ := d.docker.Source("")
		expected := d.expectedDigest
		if got != expected {
			t.Errorf("Docker Image %v:%v expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}
