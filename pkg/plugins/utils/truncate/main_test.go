package truncate

import (
	"testing"
)

func TestString(t *testing.T) {
	str := String("This is a really long PR body", 24)

	if got, want := str, "This is a really long PR"; got != want {
		t.Fatalf(`str = %q, want %q`, got, want)
	}
}
