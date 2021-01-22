package file

import (
	"fmt"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {

	data, err := Read(f.File, workingDir)
	if err != nil {
		return "", err
	}

	fmt.Printf("\u2714 Content:\n%v\n\n found from file %v \n",
		Show(string(data)),
		f.File)

	return string(data), nil
}
