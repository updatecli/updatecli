package udash

import (
	"encoding/json"
	"fmt"
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
	Token string `json:"token,omitempty"`
	// API stores the api URL
	API string `json:"api,omitempty"`
	// URL stores the front URL
	URL string `json:"url,omitempty"`
}

// readConfigFile reads the config file
func readConfigFile() (*spec, error) {

	configFile, err := initConfigFile()
	if err != nil {
		return nil, fmt.Errorf("init Updatecli configuration file: %w", err)
	}

	if _, err := os.Stat(configFile); err != nil {
		return nil, fmt.Errorf("config file %s does not exist: %w", configFile, err)
	}

	configContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read Updatecli configuration file: %w", err)
	}

	data := spec{}

	if err := json.Unmarshal(configContent, &data); err != nil {
		return nil, fmt.Errorf("unmarshal Updatecli configuration file: %w", err)
	}

	return &data, nil
}

// writeConfigFile writes the config file
func writeConfigFile(configFileName string, data *spec) error {
	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal Updatecli configuration file: %w", err)
	}

	// create file
	f, err := os.Create(configFileName)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("create Updatecli configuration file: %w", err)
	}
	defer f.Close()

	// write bytes to the file
	_, err = f.Write(d)
	if err != nil {
		return fmt.Errorf("write Updatecli configuration file: %w", err)
	}
	return nil
}

// updateConfigFile updates the config file
func updateConfigFile(data authData) error {

	updatecliConfigPath, err := initConfigFile()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("init Updatecli configuration file: %w", err)
	}

	d, err := readConfigFile()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read Updatecli configuration file: %w", err)
	}

	if d == nil {
		d = &spec{}
		d.Auths = make(map[string]authData)
	}

	d.Auths[sanitizeTokenID(data.API)] = authData{
		Token: data.Token,
		API:   data.API,
		URL:   data.URL,
	}
	d.Default = sanitizeTokenID(data.API)

	err = writeConfigFile(updatecliConfigPath, d)
	if err != nil {
		return fmt.Errorf("write Updatecli configuration file: %w", err)
	}

	return nil
}

// ConfigFilePath returns the path of the config file
func ConfigFilePath() (string, error) {
	configFile, err := initConfigFile()
	if err != nil {
		return "", fmt.Errorf("init Updatecli configuration file: %w", err)
	}

	// Testing if configFile exists
	if _, err = os.Open(configFile); err != nil {
		return "", fmt.Errorf("config file %s does not exist: %w", configFile, err)
	}

	return configFile, nil
}
