package commentmap

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	comments, err := Get("")

	if err != nil {
		t.Errorf("Unexpected Error: %v", err)

	}

	expectedResult := false

	for key := range comments {
		if strings.Contains(key, "github.com/updatecli/updatecli/") {
			expectedResult = true
			break
		}
	}

	if !expectedResult {
		t.Errorf("Unexpected result: %v", comments)
	}

}
