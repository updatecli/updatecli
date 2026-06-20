package config

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// getFieldValueByQuery returns the string value located at the given query path.
func getFieldValueByQuery(conf interface{}, query []string) (string, error) {
	field, err := lookupFieldByQuery(conf, query)
	if err != nil {
		return "", err
	}
	return field.String(), nil
}

// getFieldBoolByQuery returns the boolean value located at the given query path.
func getFieldBoolByQuery(conf interface{}, query []string) (bool, error) {
	field, err := lookupFieldByQuery(conf, query)
	if err != nil {
		return false, err
	}
	if field.Kind() != reflect.Bool {
		return false, fmt.Errorf("field %q is not a boolean", strings.Join(query, "."))
	}
	return field.Bool(), nil
}

// lookupFieldByQuery traverses conf following the case-insensitive query path
// and returns the resolved reflect.Value.
func lookupFieldByQuery(conf interface{}, query []string) (reflect.Value, error) {
	if query == nil {
		query = make([]string, 0)
	}

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
		// Field = ValueIface.Elem().FieldByName(query[0])

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
	}

	// Means that despite the different case sensitive key, we couldn't find it
	if !Field.IsValid() {
		logrus.Debugf(
			"Configuration `%s` does not have the field `%s`",
			ValueIface.Type(),
			query[0])
		return reflect.Value{}, ErrNoKeyDefined
	}

	if len(query) > 1 {
		return lookupFieldByQuery(Field.Interface(), query[1:])
	}

	return Field, nil
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
