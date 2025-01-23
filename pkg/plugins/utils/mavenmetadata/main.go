package mavenmetadata

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	"golang.org/x/text/encoding/ianaindex"
)

// DefaultHandler is the default implementation for a maven metadata handler
type DefaultHandler struct {
	metadataURL string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter    version.Filter
	contentRetriever text.TextRetriever
}

// New returns a newly initialized DefaultHandler object
func New(metadataURL string, versionFilter version.Filter) *DefaultHandler {
	if versionFilter.IsZero() {
		versionFilter.Kind = "latest"
	}

	return &DefaultHandler{
		metadataURL:      metadataURL,
		versionFilter:    versionFilter,
		contentRetriever: &text.Text{},
	}
}

// getMetadataFile is an internal method that returns the parsed metadata object
func (d *DefaultHandler) getMetadataFile() (metadata, error) {
	body, err := d.contentRetriever.ReadAll(d.metadataURL)
	if err != nil {
		return metadata{}, err
	}
	data := metadata{}
	decoder := xml.NewDecoder(bytes.NewBuffer([]byte(body)))

	decoder.CharsetReader = func(charset string, reader io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, err
		}
		if enc == nil {
			return nil, fmt.Errorf("no decoder found for: %s", charset)
		}
		return enc.NewDecoder().Reader(reader), nil
	}

	if err := decoder.Decode(&data); err != nil {
		return metadata{}, err
	}

	return data, nil
}

func (d *DefaultHandler) GetLatestVersion() (string, error) {
	data, err := d.getMetadataFile()
	if err != nil {
		return "", err
	}

	switch d.versionFilter.Kind {
	case "latest", "":
		if data.Versioning.Latest == "" {
			return "", fmt.Errorf("%s No latest version found at %s", result.FAILURE, redact.URL(d.metadataURL))
		}
		return data.Versioning.Latest, nil

	default:
		versions := []string{}
		versions = append(versions, data.Versioning.Versions.Version...)

		v, err := d.versionFilter.Search(versions)
		if err != nil {
			return "", err
		}

		return v.GetVersion(), nil
	}
}

func (d *DefaultHandler) GetVersions() ([]string, error) {
	data, err := d.getMetadataFile()
	if err != nil {
		return []string{}, err
	}

	versions := []string{}
	versions = append(versions, data.Versioning.Versions.Version...)

	return versions, nil
}

func (d *DefaultHandler) GetMetadataURL() string {
	return d.metadataURL
}
