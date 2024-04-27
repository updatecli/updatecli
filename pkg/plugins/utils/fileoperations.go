package utils

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
)

func Base64EncodeFile(path string) (string, error) {
	in, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer in.Close()

	buf := bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	if _, err := io.Copy(encoder, in); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
