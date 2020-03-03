package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
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

// ReadFile is just a abstraction in front of ReadYamlFile or ReadTomlFile
func (config *Config) ReadFile(cfgFile string) {

	v := viper.New()

	v.SetEnvPrefix("updatecli")
	v.AutomaticEnv()
	v.SetConfigName("updateCli")        // name of config file (without extension)
	v.SetConfigType("yaml")             // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath("$HOME/.updateCli") // call multiple times to add many search paths
	v.AddConfigPath(".")                // optionally look for config in the working directory
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found")
		} else {
			fmt.Println(err)
		}
	}

	err := v.Unmarshal(&config)
	if err != nil {
		fmt.Printf("unable to decode into struct, %v\n", err)
	}
}

// Check is a function to test that some settings are correctly present
func (config *Config) Check() bool {
	return true
}

// Display show loaded configuration
func (config *Config) Display() {
	fmt.Println(config)
}
