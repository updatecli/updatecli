package truncate

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	// small string
	smallStr := "This is a small Pull Request body"
	truncatedSmallStr := "This is a small Pull Request"

	// too large string
	const REPEAT = 2500
	str := "This is a really long Pull Request body"

	longStr := ""
	for i := 0; i < REPEAT; i++ {
		longStr += str
	}

	truncatedLongStr := ""
	for i := 0; i < REPEAT-1; i++ {
		truncatedLongStr += str
	}

	// execute test cases
	tests := []struct {
		s             string
		truncatedSLen int
		want          string
	}{
		{smallStr, len(truncatedSmallStr), truncatedSmallStr},
		{longStr, len(truncatedLongStr), truncatedLongStr},
	}

	for i, test := range tests {
		testname := fmt.Sprintf("test case: %d, length: %d", i+1, test.truncatedSLen)
		t.Run(testname, func(t *testing.T) {
			res := String(test.s, test.truncatedSLen)
			if res != fmt.Sprint(test.want, "...") {
				t.Fatalf(`str = %q, want %q`, res, test.want)
			}
		})
	}
}
