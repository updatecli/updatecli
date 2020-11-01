package tmp

import (
	"os"
	"path"
)

var (
	//Directory defines where updatecli will save temporary files like git clone.
	Directory = path.Join(os.TempDir(), "updatecli")
)

// Clean will remove the main temporary directory used by updatecli.
func Clean() error {
	err := os.RemoveAll(Directory)

	if err != nil {
		return err
	}

	return nil
}

// Create will create the main temporary directory used by updatecli
func Create() error {

	if _, err := os.Stat(Directory); os.IsNotExist(err) {

		err := os.MkdirAll(Directory, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
