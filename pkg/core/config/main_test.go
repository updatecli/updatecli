package config

import (
	"testing"
)

func TestChecksum(t *testing.T) {
	got, err := Checksum("./checksum.example")
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	if err != nil {
		t.Errorf("Got an unexpected error: %q", err.Error())
	}

	if got != expected {
		t.Errorf("Got %q, expected %q", got, expected)
	}
}
