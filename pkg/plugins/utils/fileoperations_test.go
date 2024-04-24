package utils

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64EncodeFile(t *testing.T) {
	fileContent := "test file content"
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(fileContent)
	if err != nil {
		t.Fatal(err)
	}

	encodedString, err := Base64EncodeFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		t.Fatal(err)
	}

	decodedString := string(decodedBytes)

	assert.Equal(t, fileContent, decodedString, "Decoded string does not match original content")
}
