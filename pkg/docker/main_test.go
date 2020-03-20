package docker

import "testing"

func TestParameters(t *testing.T) {
	d := &Docker{
		Image: "nginx",
	}

	ok, _ := d.Check()

	// Test if current setting is valid
	if ok != true {
		t.Errorf("Minimum valid configuration provided! Expect %v, got %v", true, ok)
	}

	// Test if we correctly return library images if dockerhub namespace is not specified
	got := d.Image
	expected := "library/nginx"
	if got != expected {
		t.Errorf("Image is configured without namespace! expected %v, got %v", expected, got)
	}

	// Test if we correctly return the default docker hub url if not defined
	expected = "hub.docker.com"
	got = d.URL
	if got != expected {
		t.Errorf("URL is not configured! expected value %v, got %v", expected, got)
	}

	// Test if we correctly return the default architecture if not defined
	expected = "amd64"
	got = d.Architecture
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

func TestIsPublished(t *testing.T) {
	// Test if existing image tag return true
	d := &Docker{
		URL:   "hub.docker.com",
		Tag:   "latest",
		Image: "olblak/updatecli",
	}

	got := d.IsTagPublished()
	expected := true
	if got != expected {
		t.Errorf("%v:%v is published! expected %v, got %v", d.Image, d.Tag, expected, got)
	}

	// Test if none existing image tag return false
	d = &Docker{
		URL:   "hub.docker.com",
		Tag:   "donotexist",
		Image: "olblak/updatecli",
	}

	got = d.IsTagPublished()
	expected = false
	if got != expected {
		t.Errorf("%v:%v is not published! expected %v, got %v", d.Image, d.Tag, expected, got)
	}

}

func TestSource(t *testing.T) {
	// Test if existing return the correct digest
	d := &Docker{
		URL:   "hub.docker.com",
		Tag:   "v0.0.2",
		Image: "olblak/updatecli",
	}

	got, _ := d.Source()
	expected := "4f9936580d3caa6b7a27da62df78acf0294277a4b62bc128de7b88ff836ed2a9"

	if got != expected {
		t.Errorf("Docker Image %v:%v expect digest %v, got %v", d.Image, d.Tag, expected, got)
	}

	// Test if non existing tag return empty string
	d = &Docker{
		URL:   "hub.docker.com",
		Tag:   "donotexist",
		Image: "olblak/updatecli",
	}
	got, _ = d.Source()
	expected = ""
	if got != expected {
		t.Errorf("Digest expected %v, got %v", expected, got)
	}
}
