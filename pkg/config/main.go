package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Config hold our cli configuration
type Config struct {
	Source     Source
	Conditions []Condition
	Targets    []Target
}

// Source define...
type Source struct {
	Kind   string
	Output string
	Spec   interface{}
}

// Condition define...
type Condition struct {
	Name string
	Kind string
	Spec interface{}
}

// Target define ...
type Target struct {
	Name       string
	Kind       string
	Spec       interface{}
	Repository interface{}
}

// ReadTomlFile read settings from a toml file
func (config *Config) ReadTomlFile(cfgFile string) {

	if _, err := toml.DecodeFile(cfgFile, &config); err != nil {
		log.Println(err)
		return
	}
}

// ReadYamlFile read settings from a yaml file
func (config *Config) ReadYamlFile(cfgFile string) {
	file, err := os.Open(cfgFile)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// ReadFile is just a abstraction in front of ReadYamlFile or ReadTomlFile
func (config *Config) ReadFile(cfgFile string) {

	config.ReadYamlFile(cfgFile)

}

// Check is a function to test that some settings are correctly present
func (config *Config) Check() bool {
	return true
}

// Display show loaded configuration
func (config *Config) Display() {
	fmt.Println(config)
}
