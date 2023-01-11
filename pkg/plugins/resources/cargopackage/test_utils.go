package cargopackage

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateDummyIndex() (string, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	index, err := os.Create(filepath.Join(dir, "config.json"))
	if err != nil {
		return "", err
	}
	defer index.Close()
	_, err = fmt.Fprintf(index, "{\"dl\":\"https://example.com\"}")
	if err != nil {
		return "", err
	}
	crateDir := filepath.Join(dir, "cr/at")
	err = os.MkdirAll(crateDir, 0750)
	if err != nil {
		return "", err
	}
	crateFile, err := os.Create(filepath.Join(crateDir, "crate-test"))
	if err != nil {
		return "", err
	}
	defer crateFile.Close()
	_, err = fmt.Fprintf(crateFile, "{\"name\":\"crate-test\",\"vers\":\"0.1.0\",\"deps\":[],\"features\":{},\"cksum\":\"b274d286f7a6aad5a7d5b5407e9db0098c94711fb3563bf2e32854a611edfb63\",\"yanked\":false,\"links\":null}")
	if err != nil {
		return "", err
	}
	return dir, nil
}
