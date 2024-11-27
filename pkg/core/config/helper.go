package config

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"cuelang.org/go/cue/cuecontext"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// FileChecksum returns sha256 checksum based on a file content.
func FileChecksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Debugf("Can't open file %q", filename)
		return "", err
	}

	defer file.Close()
	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// Checksum returns sha256 checksum based on a file content.
func Checksum(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}

// IsTemplatedString test if a string contains go template information
func IsTemplatedString(s string) bool {
	if len(s) == 0 {
		return false
	}

	leftDelimiterFound := false

	for _, val := range strings.SplitAfter(s, "{{") {
		if strings.Contains(val, "{{") {
			leftDelimiterFound = true
			continue
		}
		if strings.Contains(val, "}}") && leftDelimiterFound {
			return true
		}
	}

	return false
}

func getFieldByQuery(conf interface{}, query []string) (value reflect.Value, err error) {
	ValueIface := reflect.ValueOf(conf)

	Field := reflect.Value{}
	titleCaser := cases.Title(language.English, cases.NoLower)

	// We want to be able to use case insensitive key
	insensitiveQuery := []string{
		query[0],
		strings.ToLower(query[0]),
		titleCaser.String(strings.ToLower(query[0])),
		titleCaser.String(query[0]),
		strings.ToTitle(query[0]),
	}

	switch ValueIface.Kind() {
	case reflect.Ptr:
		// Check if the passed interface is a pointer
		// Create a new type of Iface's Type, so we have a pointer to work with
		// 'dereference' with Elem() and get the field by name
		//Field = ValueIface.Elem().FieldByName(query[0])

		for _, q := range insensitiveQuery {
			Field = ValueIface.Elem().FieldByName(q)
			if Field.IsValid() {
				query[0] = q
				break
			}
		}
	case reflect.Map:
		// We want to be able to use case insensitive key
		for _, q := range insensitiveQuery {
			Field = ValueIface.MapIndex(reflect.ValueOf(q))
			if Field.IsValid() {
				query[0] = q
				break
			}
		}
	case reflect.Struct:
		// We want to be able to use case insensitive key
		for _, q := range insensitiveQuery {
			Field = ValueIface.FieldByName(q)
			if Field.IsValid() {
				break
			}
		}
	case reflect.Slice:
		// Handle slice: Get the first element's "Value" field if it exists
		index, err := strconv.Atoi(query[0])
		if err != nil {
			return value, fmt.Errorf("Could not use %q as slice index: %s", query[0], err)
		}
		if index >= ValueIface.Len() {
			return value, fmt.Errorf("Could not use %q as slice index: not enough elem in slice", query[0])
		}
		Field = ValueIface.Index(index)
		if Field.IsValid() {
			break
		}
	}

	// Means that despite the different case sensitive key, we couldn't find it
	if !Field.IsValid() {
		logrus.Debugf(
			"Configuration `%s` does not have the field `%s`",
			ValueIface.Type(),
			query[0])
		return value, ErrNoKeyDefined
	}

	if len(query) > 1 {
		value, err = getFieldByQuery(Field.Interface(), query[1:])
		if err != nil {
			return value, err
		}

	} else if len(query) == 1 {
		return Field, nil
	}

	return value, nil

}
func getFieldValueByQuery(conf interface{}, query []string) (string, error) {
	field, err := getFieldByQuery(conf, query)
	if err != nil {
		return "", err
	}
	return field.String(), nil
}

// readCueConfig loads a cue spec and convert it to YAML before converting it to an Updatecli config spec
// An important limitation in today's Updatecli implementation is that
// Updatecli loads all configuration in memory and then apply each files individually as independent pipeline.
// So cuelang feature won't be able to load module or package using the directory structure.
func readCueConfig(in []byte) ([]byte, error) {

	ctx := cuecontext.New()

	compiledVal := ctx.CompileBytes(in)
	if compiledVal.Err() != nil {
		return nil, fmt.Errorf("compile cue spec: %w", compiledVal.Err())
	}

	val, err := cueyaml.Encode(compiledVal)
	if err != nil {
		return nil, fmt.Errorf("encode cue spec to yaml: %w", err)
	}

	return val, nil
}

// unmarshalConfigSpec unmarshal an Updatecli config spec
func unmarshalConfigSpec(in []byte, out *[]Spec) error {

	r := bytes.NewReader(in)
	dec := yaml.NewDecoder(r)

	for {
		var s Spec
		if err := dec.Decode(&s); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		*out = append(*out, s)
	}

	return nil
}
