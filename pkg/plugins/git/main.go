package git

import (
	"fmt"
	"os"
	"path"

	"github.com/olblak/updateCli/pkg/core/tmp"
)

// Git contains settings to manipulate a git repository.
type Git struct {
	URL          string
	Username     string
	Password     string
	Branch       string
	remoteBranch string
	User         string
	Email        string
	Directory    string
	Version      string
}

func (g *Git) setDirectory() {

	directory := path.Join(tmp.Directory, g.URL)

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	g.Directory = directory

	fmt.Printf("Directory: %v\n", g.Directory)
}
