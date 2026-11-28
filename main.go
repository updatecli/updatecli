package main

//go:generate go run ./winres/gen.go

import (
	"github.com/updatecli/updatecli/cmd"
)

func main() {
	cmd.Execute()
}
