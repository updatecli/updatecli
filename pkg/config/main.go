package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/helm"
	"gopkg.in/yaml.v2"
)

var (
	configFileName string = "updateCli.yaml"
	configFilePath string = "."
)

// Config hold our cli configuration
type Config struct {
	Github github.Github
	Docker docker.Docker
	Helm   helm.Helm
}

// ReadTomlFile read settings from a toml file
func (config *Config) ReadTomlFile() {

	if _, err := toml.DecodeFile(filepath.Join(configFilePath, configFileName), &config); err != nil {
		log.Println(err)
		return
	}
}

// ReadYamlFile read settings from a yaml file
func (config *Config) ReadYamlFile() {
	file, err := os.Open(filepath.Join(configFilePath, configFileName))
	defer file.Close()
	if err != nil {
		fmt.Println(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
	}

}

// ReadFile is just a abstraction in front of ReadYamlFile or ReadTomlFile
func (config *Config) ReadFile() {

	config.ReadYamlFile()

}

// Check is a function to test that some settings are correctly present
func (config *Config) Check() bool {
	if config.Docker.Tag == "" {
		log.Printf("No docker image tag specified at %s/%s, so trying to guess based on release drafter information\n",
			configFilePath,
			configFileName)
		config.Docker.Tag = config.Github.GetVersion()
		return false
	}
	return true
}

// Display show loaded configuration
func (config *Config) Display() {
	fmt.Println(config)
}
