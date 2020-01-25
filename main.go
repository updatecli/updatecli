package main

import (
	"github.com/olblak/updateCli/pkg/config"
)

var conf config.Config

func init() {
	conf.ReadFile()
	conf.Check()
}

func main() {

	conf.Display()
	conf.Helm.UpdateChart(conf.Github.GetVersion())
}
