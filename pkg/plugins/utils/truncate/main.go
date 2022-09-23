package truncate

import (
	"fmt"
	"unicode/utf8"
)

// String safely truncate a string
func String(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return fmt.Sprint(string([]rune(str)[:length]), "...")
}
