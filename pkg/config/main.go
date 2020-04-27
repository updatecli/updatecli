package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olblak/updateCli/pkg/engine/condition"
	"github.com/olblak/updateCli/pkg/engine/source"
	"github.com/olblak/updateCli/pkg/engine/target"
	"github.com/spf13/viper"
)

// Config contains cli configuration
type Config struct {
	Source     source.Source
	Conditions map[string]condition.Condition
	Targets    map[string]target.Target
}

// Reset reset configuration
func (config *Config) Reset() {
	config.Source = source.Source{}
	config.Conditions = map[string]condition.Condition{}
	config.Targets = map[string]target.Target{}
}

// ReadFile reads the updatecli configuration file
func (config *Config) ReadFile(cfgFile string) {

	config.Reset()

	dirname, basename := filepath.Split(cfgFile)

	switch extension := filepath.Ext(basename); extension {
	case ".tpl", ".tmpl":
		t := Template{
			CfgFile:    filepath.Join(dirname, basename),
			ValuesFile: "values.yaml",
		}

		err := t.Unmarshal(config)
		if err != nil {
			fmt.Println(err)
		}

	case ".yaml", ".yml":
		v := viper.New()

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
	default:
		fmt.Printf("File extension not supported: %v", extension)
	}

	os.Exit(32)

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
