package temurin

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Temurin) Source(workingDir string, resultSource *result.Source) error {
	// Start by getting the version (required in any case)
	releaseNames, err := t.apiGetReleaseNames()
	if err != nil {
		resultSource.Result = result.FAILURE
		return err
	}

	if len(releaseNames) == 0 {
		logrus.Debug("[temurin] empty response for 'release_names'.")
		resultSource.Result = result.FAILURE
		return fmt.Errorf("[temurin] No release found matching provided criteria. Use '--debug' to get details.")
	}

	// Only get the most recent, e.g. the first one (DESC order)
	t.foundVersion = releaseNames[0]

	switch t.spec.Result {
	case "version":
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found version %q", t.foundVersion)
		resultSource.Information = t.foundVersion
		return nil

	case "installer_url":
		installerUrl, err := t.apiGetInstallerUrl(t.foundVersion)
		if err != nil {
			resultSource.Result = result.FAILURE
			return err
		}
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found installer URL %q (version %q)", installerUrl, t.foundVersion)
		resultSource.Information = installerUrl
		return nil

	case "checksum_url":
		installerChecksumUrl, err := t.apiGetChecksumUrl(t.foundVersion)
		if err != nil {
			resultSource.Result = result.FAILURE
			return err
		}
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found installer checksum URL %q (version %q)", installerChecksumUrl, t.foundVersion)
		resultSource.Information = installerChecksumUrl
		return nil

	case "signature_url":
		signatureUrl, err := t.apiGetSignatureUrl(t.foundVersion)
		if err != nil {
			resultSource.Result = result.FAILURE
			return err
		}
		resultSource.Result = result.SUCCESS
		resultSource.Description = fmt.Sprintf("[temurin] found installer signature URL %q (version %q)", signatureUrl, t.foundVersion)
		resultSource.Information = signatureUrl
		return nil

	default:
		return fmt.Errorf("[temurin] Unknown value %q for 'result' field.", t.spec.Result)

	}
}
