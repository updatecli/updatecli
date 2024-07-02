package toolversions

// Put inserts or updates a key-value pair in the .tool-versions file
func (f *FileContent) Put(key, value string) error {
	found := false
	for i, entry := range f.Entries {
		if entry.Key == key {
			f.Entries[i].Value = value // Update existing entry
			found = true
			break
		}
	}
	if !found {
		f.Entries = append(f.Entries, Entry{Key: key, Value: value}) // Add new entry
	}

	return nil
}
