package docker

import "testing"

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

func TestGetVersion(t *testing.T) {
	// Test if existing return the correct digest
	d := &Docker{
		URL:   "hub.docker.com",
		Tag:   "latest",
		Image: "olblak/updatecli",
	}
	got := d.GetVersion()
	expected := "sha256:535c6eda6ce32e8c3309878bd27faa0cd41c0cb833149bf5544c7bccff817541"
	if got != expected {
		t.Errorf("Digest expected %v, got %v", expected, got)
	}

	// Test if non existing tag return empty string
	d = &Docker{
		URL:   "hub.docker.com",
		Tag:   "donotexist",
		Image: "olblak/updatecli",
	}
	got = d.GetVersion()
	expected = ""
	if got != expected {
		t.Errorf("Digest expected %v, got %v", expected, got)
	}
}
