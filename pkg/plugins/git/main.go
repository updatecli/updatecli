package git

import (
	"fmt"
	"os"
	"path"
	"strings"

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

	URL := strings.Replace(g.URL, "/", "_", -1)
	URL = strings.Replace(URL, "\\", "_", -1)
	URL = strings.Replace(URL, " ", "_", -1)

	directory := path.Join(tmp.Directory, URL)

	for _, dir := range []string{tmp.Directory, URL} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {

			err := os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	g.Directory = directory

	fmt.Printf("Directory: %v\n", g.Directory)
}
