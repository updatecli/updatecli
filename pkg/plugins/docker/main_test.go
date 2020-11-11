package docker

import (
	"fmt"
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
			Tag:   "latest",
		},
		expectedCondition: true,
		expectedHostname:  "hub.docker.com",
		expectedImage:     "library/nginx",
		expectedDigest:    "34f3f875e745861ff8a37552ed7eb4b673544d2c56c7cc58f9a9bec5b4b3530e",
	},
	{
		docker: Docker{
			Image: "mcr.microsoft.com/azure-cli",
			Tag:   "latest",
		},
		expectedCondition: true,
		expectedHostname:  "mcr.microsoft.com",
		expectedImage:     "azure-cli",
		expectedDigest:    "bddcbadc711fd3c0a41c3101a0ba07ace4c1b124b6ee7b57be3ffe8142f140c9",
	},
	{
		docker: Docker{
			Image: "ghcr.io/olblak/updatecli",
			Tag:   "v0.0.22",
		},
		expectedCondition: true,
		expectedHostname:  "ghcr.io",
		expectedImage:     "olblak/updatecli",
		expectedDigest:    "xxx",
	},
	{
		docker: Docker{
			Image: "quay.io/jetstack/cert-manager-controller",
			Tag:   "v1.0.0",
		},
		expectedCondition: true,
		expectedHostname:  "quay.io",
		expectedImage:     "jetstack/cert-manager-controller",
		expectedDigest:    "xxx",
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

	// Test if we correctly return the default architecture if not defined
	expected := "amd64"
	got := d.Architecture
	if got != expected {
		t.Errorf("Architecture is not configured! expected value %v, got %v", expected, got)
	}

	// Test if we correctly return the default docker hub url if not defined
	expected = "latest"
	got = d.Tag
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
		got, _ := d.docker.Source()
		expected := d.expectedDigest
		if got != expected {
			t.Errorf("Docker Image %v:%v expect digest %v, got %v", d.docker.Image, d.docker.Tag, expected, got)
		}
	}
}
