package udash

import (
	"encoding/json"
	"errors"
	"os"
)

// spec defines the structure of the config file
type spec struct {
	// Auths stores the authentication data
	Auths map[string]authData
	// Default stores the default authentication data
	Default string
}

// authData defines the structure of the authentication data
type authData struct {
	// Token stores the access token
	Token string
	// Api stores the api URL
	Api string
	// URL stores the front URL
	URL string
}

// readConfigFile reads the config file
func readConfigFile() (*spec, error) {

	configFile, err := initConfigFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	configContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	data := spec{}

	if err := json.Unmarshal(configContent, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// writeConfigFile writes the config file
func writeConfigFile(configFileName string, data *spec) error {
	d, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	// create file
	f, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// write bytes to the file
	_, err = f.Write(d)
	if err != nil {
		return err
	}
	return nil
}

// ConfigFilePath returns the path of the config file
func ConfigFilePath() (string, error) {
	configFile, err := initConfigFile()
	if err != nil {
		return "", err
	}

	// Testing if configFile exists
	if _, err = os.Open(configFile); err != nil {
		return "", err
	}

	return configFile, nil
}
