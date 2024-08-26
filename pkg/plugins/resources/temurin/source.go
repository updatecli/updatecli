package temurin

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Temurin) Source(workingDir string, resultSource *result.Source) error {
	// Start by getting the version (required in any case)
	releaseName, err := t.apiGetReleaseName()
	if err != nil {
		resultSource.Result = result.FAILURE
		return err
	}

	t.foundVersion = releaseName

	switch t.spec.Result {
	case "version":
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found version %q", releaseName)
		resultSource.Information = releaseName
		return nil

	case "installer_url":
		installerUrl, err := t.apiGetInstallerUrl(releaseName)
		if err != nil {
			resultSource.Result = result.FAILURE
			return err
		}
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found installer URL %q (version %q)", installerUrl, releaseName)
		resultSource.Information = installerUrl
		return nil

	case "checksum_url":
		installerChecksumUrl, err := t.apiGetChecksumUrl(releaseName)
		if err != nil {
			resultSource.Result = result.FAILURE
			return err
		}
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found installer checksum URL %q (version %q)", installerChecksumUrl, releaseName)
		resultSource.Information = installerChecksumUrl
		return nil

	case "signature_url":
		signatureUrl, err := t.apiGetSignatureUrl(releaseName)
		if err != nil {
			resultSource.Result = result.FAILURE
			return err
		}
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found installer signature URL %q (version %q)", signatureUrl, releaseName)
		resultSource.Information = signatureUrl
		return nil

	default:
		return fmt.Errorf("[temurin] Unknown value %q for 'result' field.", t.spec.Result)

	}
}
