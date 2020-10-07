package dockerfile

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

var (
	rawDockerfile string = `
		---
		FROM ubuntu:20.04
		#Comments
		RUN echo "Hello World"
	`
)

func TestSearchNode(t *testing.T) {
	fmt.Println(rawDockerfile)

	if true {
		t.Errorf("Debug")
	}

}

func TestShow(t *testing.T) {

	data, err := parser.Parse(bytes.NewReader([]byte(rawDockerfile)))

	for position, line := range rawDockerfile {
		fmt.Printf("%v - %s\n", position, string(line))
	}

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(rawDockerfile)
	fmt.Println("++++++++++++++")
	fmt.Println(show(data.AST))
	fmt.Println("++++++++++++++")
	fmt.Println(data.AST.Dump())

	fmt.Println("++++++++++++++")

	fmt.Println(lines(data.AST))
	fmt.Println("++++++++++++++")

	if true {
		t.Errorf("Debug")
	}

}
