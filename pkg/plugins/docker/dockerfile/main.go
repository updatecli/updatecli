package dockerfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

//	https://github.com/moby/buildkit/blob/master/frontend/dockerfile/parser/parser
// https://github.com/moby/buildkit/issues/1561

// Dockerfile is struct that contains parametes to interact with a Dockerfile
type Dockerfile struct {
	Path        string
	File        string
	Instruction string
	Value       string
	DryRun      bool
}

// ReadFile read a Dockerfile
func (d *Dockerfile) ReadFile() (data []byte, err error) {
	path := filepath.Join(d.Path, d.File)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	data, err = ioutil.ReadAll(file)

	if err != nil {
		return nil, err
	}

	return data, err
}

func lines(node *parser.Node) (result []int) {

	for _, n := range node.Children {
		fmt.Println(n.StartLine)
		result = append(result, n.StartLine)
	}

	for n := node.Next; n != nil; n = n.Next {
		fmt.Println(n.StartLine)
		result = append(result, n.StartLine)
	}

	return result
}

func show(node *parser.Node) string {
	str := ""
	str += node.Value

	if len(node.Flags) > 0 {
		str += fmt.Sprintf(" %q", node.Flags)
	}

	for _, n := range node.Children {
		str += n.Dump() + "\n"
	}

	for n := node.Next; n != nil; n = n.Next {
		if len(n.Children) > 0 {
			str += " " + n.Dump()
		} else {
			str += " " + n.Value
		}
	}

	return strings.TrimSpace(str)
}

func (d *Dockerfile) search(node *parser.Node) (found bool, err error) {

	fmt.Printf("Instruction: %v\n", node.Value)

	if len(node.Children) > 0 {
		for i, child := range node.Children {
			fmt.Printf("Exploring child %v\n", i)
			found, err = d.search(child)
			if found {
				return found, nil
			}
		}
	}

	if node.Next == nil {
		return false, nil
	}

	fmt.Printf("Instruction: %v %v\n", node.Value, node.Next.Value)
	if strings.ToUpper(d.Instruction) == strings.ToUpper(node.Value) &&
		strings.ToUpper(d.Value) == strings.ToUpper(node.Next.Value) {
		found = true
		fmt.Println("Found")
		return true, nil
	}

	//if strings.ToUpper(d.Instruction) == "LABEL" &&
	//	len(strings.Split(d.Value, "=")) > 1 {
	//	// if label
	//	if strings.ToUpper(d.Instruction) == strings.ToUpper(node.Value) &&
	//		strings.ToUpper(d.Value) == strings.ToUpper(node.Next.Value) {
	//		found = true
	//		fmt.Println("Found")
	//		return true, nil
	//	}
	//}

	return false, err
}
