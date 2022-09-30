package dasel

// Put insert value in a Dasel node
func (f *FileContent) Put(query, value string) error {
	if f.DaselNode == nil {
		return ErrEmptyDaselNode
	}

	if err := f.DaselNode.PutMultiple(query, value); err != nil {
		return err
	}

	return nil
}

// PutMultiple insert multiple value in a Dasel node
func (f *FileContent) PutMultiple(query, value string) error {
	if f.DaselNode == nil {
		return ErrEmptyDaselNode
	}

	if err := f.DaselNode.PutMultiple(query, value); err != nil {
		return err
	}

	return nil
}
