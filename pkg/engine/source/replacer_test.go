package source

import (
	"testing"
	"reflect"
)

func TestReplacersUnmarshall(t *testing.T) {
	dataSet := Replacers{
		{
			Source: "a",
			Destination: "b",
		},
		{
			Source: "c",
			Destination: "d",
		},
		{
			Source: "e",
			Destination: "f",
		},
	}
	expected := []string{ "a","b","c","d","e","f"}

	got := dataSet.Unmarshal()

	// Testing that we correctly get a slice of string
	if ! reflect.DeepEqual(expected, got) {
		t.Errorf("List of replacers is incorrect, expected '%v', got '%v'", expected, got)
	}


	// Testing that order matter
	expected = []string{ "a","c","b","d","e","d"}

	if reflect.DeepEqual(expected, got) {
		t.Errorf("List of replacers is incorrect, expected '%v', got '%v'", expected, got)
	}
}