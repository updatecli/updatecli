package dasel

import (
	"fmt"

	daselV2 "github.com/tomwright/dasel/v2"
)

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

func (f *FileContent) PutV2(query, value string) error {
	if f.DaselV2Node == nil {
		return ErrEmptyDaselNode
	}

	if _, err := daselV2.Put(f.DaselV2Node, query, value); err != nil {
		return fmt.Errorf("setting value %q with query %q: %w", value, query, err)
	}

	return nil

}
