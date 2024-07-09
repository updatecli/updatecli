package toolversions

import "fmt"

// Query returns the value for a specific value from a .tool-versions file
func (f *FileContent) Get(key string) (string, error) {
	if f.Entries == nil {
		return "", ErrToolVersionsFailedParsingByteFormat
	}

	var value string
	for _, entry := range f.Entries {
		if entry.Key == key {
			value = entry.Value
			break
		}
	}

	if value == "" {
		err := fmt.Errorf("could not find value for key %q from file %q",
			key,
			f.FilePath)
		return "", err
	}

	return value, nil
}
