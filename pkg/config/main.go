package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config contains cli configuration
type Config struct {
	Source     Source
	Conditions map[string]Condition
	Targets    map[string]Target
}

// Source defines how a value is retrieved from a specific source
type Source struct {
	Kind    string
	Output  string
	Prefix  string
	Postfix string
	Spec    interface{}
}

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	Name string
	Kind string
	Spec interface{}
}

// Target defines which file need to be updated based on source output
type Target struct {
	Name       string
	Kind       string
	Spec       interface{}
	Repository interface{}
}

// Reset reset configuration
func (config *Config) Reset() {
	config.Source = Source{}
	config.Conditions = map[string]Condition{}
	config.Targets = map[string]Target{}
}

// ReadFile reads the updatecli configuration file
func (config *Config) ReadFile(cfgFile string) {

	config.Reset()
	v := viper.New()

	dirname, basename := filepath.Split(cfgFile)

	v.SetEnvPrefix("updatecli")
	v.AutomaticEnv()
	v.SetConfigName(strings.TrimSuffix(basename, filepath.Ext(basename))) // name of config file (without extension)
	v.SetConfigType(strings.Replace(filepath.Ext(basename), ".", "", -1)) // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath(dirname)                                              // optionally look for config in the working directory
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

// Check is a function that test if the configuration is correct
func (config *Config) Check() bool {
	fmt.Printf("TODO: Implement configuration check\n")
	return true
}

// Display shows updatecli configuration including secrets !
func (config *Config) Display() {
	fmt.Println(config)
}
