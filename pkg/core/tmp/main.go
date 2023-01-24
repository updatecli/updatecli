package tmp

import (
	"os"
	"path"
)

var (
	//Directory defines where updatecli will save temporary files like git clone.
	Directory    = path.Join(os.TempDir(), "updatecli")
	BinDirectory = path.Join(Directory, "bin")
)

// Clean removes the Updatecli temporary root directory.
func Clean() error {
	err := os.RemoveAll(Directory)

	if err != nil {
		return err
	}

	return nil
}

// Create creates Updatecli temporary directory
func Create() error {
	if _, err := os.Stat(Directory); os.IsNotExist(err) {

		err := os.MkdirAll(Directory, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// InitBin creates a bin directory used by updatecli to store and execute command
func InitBin() (string, error) {
	if _, err := os.Stat(BinDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(BinDirectory, 0755)
		if err != nil {
			return "", err
		}
	}
	return BinDirectory, nil
}
