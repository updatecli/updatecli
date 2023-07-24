package udash

import (
	"encoding/json"
	"errors"
	"os"
)

type spec struct {
	Auths   map[string]authData
	Default string
}

type authData struct {
	// Token stores the access token
	Token string
	// Api stores the api URL
	Api string
	// URL stores the front URL
	URL string
}

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
