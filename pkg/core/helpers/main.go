package helpers

import (
	"io/ioutil"
	"os"
)

// ReadFile read a file then return a array of byte
func ReadFile(file string) (data []byte, err error) {

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	data, err = ioutil.ReadAll(f)

	if err != nil {
		return nil, err
	}

	return data, err
}
