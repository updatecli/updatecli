package dockerfile

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Target update a Dockerfile field if needed
func (d *Dockerfile) Target() (changed bool, err error) {
	changed = false

	raw, err := d.ReadFile()
	keyFound := false
	if err != nil {
		return false, err
	}
	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	fmt.Printf("Dump, %v\n\n", data.AST.Dump())

	if len(data.AST.Children) > 0 {
		for _, child := range data.AST.Children {
			if data.AST.Value == d.Instruction {
				keyFound = true
			}
			if strings.ToUpper(child.Value) == strings.ToUpper(d.Instruction) &&
				strings.ToUpper(child.Next.Value) == strings.ToUpper(d.Value) {
				keyFound = true
				fmt.Printf("\u2714 Instruction '%s %s' from Dockerfile '%s', is already up to date \n",
					child.Value,
					child.Next.Value,
					d.File)
				break
			}
			if strings.ToUpper(child.Value) == strings.ToUpper(d.Instruction) &&
				strings.ToUpper(child.Next.Value) != strings.ToUpper(d.Value) {
				keyFound = true

				fmt.Printf("\u2714 Instruction '%s %s' from Dockerfile '%s', has been updated to '%s %s' \n",
					child.Value,
					child.Next.Value,
					d.File,
					child.Value,
					d.Value)

				child.Next.Value = d.Value
				changed = true

				break
			}
			fmt.Printf("%v %v\n", child.Value, child.Next.Value)
		}
	}

	if !keyFound {
		fmt.Printf("\u2717 Instruction '%s' from Dockerfile '%s', could not be found\n",
			d.Instruction,
			d.File)
	}

	//if len(data.AST.Children) > 0 {
	//	for _, child := range data.AST.Children {
	//		fmt.Println(child.Original)
	//		fmt.Printf("%s", strings.ToUpper(child.Value))
	//		if len(child.Children) > 0 {
	//			for _, child := range child.Children {
	//				fmt.Printf(" %s", child.Value)
	//			}
	//		}
	//		fmt.Printf(" %s\n", child.Next.Value)
	//	}
	//}

	return changed, nil
}
