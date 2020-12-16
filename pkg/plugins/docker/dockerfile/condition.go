package dockerfile

import (
	"bytes"
	"fmt"
	"path"

	"github.com/olblak/updateCli/pkg/core/helpers"
	"github.com/olblak/updateCli/pkg/core/scm"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(version string) (bool, error) {

	raw, err := helpers.ReadFile(d.File)
	if err != nil {
		return false, err
	}

	if len(d.Value) == 0 {
		d.Value = version
	}

	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	found, err := d.search(data.AST)

	if err != nil {
		return false, err
	}

	if found {
		fmt.Printf("\u2714 Instruction '%s' from Dockerfile '%s', is correctly set to '%s' \n",
			d.Instruction,
			d.File,
			d.Value)
		return true, nil
	}
	fmt.Printf("\u2717 Instruction '%s' from Dockerfile '%s', is incorrectly set to '%s' \n",
		d.Instruction,
		d.File,
		d.Value)

	return false, nil

}

// ConditionFromSCM run based on a file from SCM
func (d *Dockerfile) ConditionFromSCM(version string, scm scm.Scm) (bool, error) {

	raw, err := helpers.ReadFile(path.Join(scm.GetDirectory(), d.File))

	if err != nil {
		return false, err
	}

	if len(d.Value) == 0 {
		d.Value = version
	}

	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	found, err := d.search(data.AST)

	if err != nil {
		return false, err
	}

	if found {
		fmt.Printf("\u2714 Instruction '%s' from Dockerfile '%s', is correctly set to '%s' \n",
			d.Instruction,
			d.File,
			d.Value)
		return true, nil
	}
	fmt.Printf("\u2717 Instruction '%s' from Dockerfile '%s', is incorrectly set to '%s' \n",
		d.Instruction,
		d.File,
		d.Value)

	return false, nil
}
