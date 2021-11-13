package source

// 2021/01/31
// Deprecated in favor of Transformer, need to be deleted in a future release

import (
	"reflect"
	"testing"
)

func TestReplacersUnmarshall(t *testing.T) {
	dataSet := Replacers{
		{
			From: "a",
			To:   "b",
		},
		{
			From: "c",
			To:   "d",
		},
		{
			From: "e",
			To:   "f",
		},
	}
	expected := []string{"a", "b", "c", "d", "e", "f"}

	got := dataSet.Unmarshal()

	// Testing that we correctly get a slice of string
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("List of replacers is incorrect, expected '%v', got '%v'", expected, got)
	}

	// Testing that order matter
	expected = []string{"a", "c", "b", "d", "e", "d"}

	if reflect.DeepEqual(expected, got) {
		t.Errorf("List of replacers is incorrect, expected '%v', got '%v'", expected, got)
	}
}
