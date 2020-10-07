package dockerfile

import (
	"bytes"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition() (bool, error) {
	raw, err := d.ReadFile()
	//keyFound := false
	if err != nil {
		return false, err
	}
	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	fmt.Printf("Dump\n%v\n\n", show(data.AST))
	defer fmt.Printf("Dump\n%v\n\n", show(data.AST))

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

	//if len(data.AST.Children) > 0 {
	//	for _, child := range data.AST.Children {
	//		if data.AST.Value == d.Instruction {
	//			keyFound = true
	//		}
	//		if strings.ToUpper(child.Value) == strings.ToUpper(d.Instruction) &&
	//			strings.ToUpper(child.Next.Value) == strings.ToUpper(d.Value) {
	//			fmt.Printf("\u2714 Instruction '%s' from Dockerfile '%s', is correctly set to '%s' \n",
	//				d.Instruction,
	//				d.File,
	//				d.Value)
	//			return true, nil
	//		}
	//		fmt.Printf("%v %v\n", child.Value, child.Next.Value)
	//	}
	//}

	//if !keyFound {
	//	fmt.Printf("\u2717 Instruction '%s' from Dockerfile '%s', isn't found\n",
	//		d.Instruction,
	//		d.File)
	//	return false, nil
	//}

	// fmt.Printf("\u2717 Instruction '%s' from Dockerfile '%s', is incorrectly set to '%s' \n",
	// 	d.Instruction,
	// 	d.File,
	// 	d.Value)

	// fmt.Printf("Dump, %v\n\n", data.AST.Dump())

	return false, nil

}
