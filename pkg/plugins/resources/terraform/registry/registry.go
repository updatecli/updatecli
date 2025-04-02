package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

type registryAddress struct {
	registryType string
	provider     terraformRegistryAddress.Provider
	module       terraformRegistryAddress.Module
	wellKnown    wellKnownResponse
}

type wellKnownResponse struct {
	ModulesPath  string `json:"modules.v1"`
	ProviderPath string `json:"providers.v1"`
}

func newRegistryAddress(webClient httpclient.HTTPClient, spec Spec) (registryAddress, error) {
	if spec.RawString == "" {
		for i, s := range []string{spec.Hostname, spec.Namespace, spec.Name, spec.TargetSystem} {
			if len(s) > 0 {
				if i > 0 && len(spec.RawString) > 0 {
					spec.RawString += "/"
				}
				spec.RawString += s
			}
		}
	}

	registryAddress := registryAddress{}

	registryAddress.registryType = spec.Type

	if spec.Type == TypeProvider {
		provider, err := terraformRegistryAddress.ParseProviderSource(spec.RawString)
		if err != nil {
			return registryAddress, err
		}

		registryAddress.provider = provider
	}

	if spec.Type == TypeModule {
		module, err := terraformRegistryAddress.ParseModuleSource(spec.RawString)
		if err != nil {
			return registryAddress, err
		}

		registryAddress.module = module
	}

	err := registryAddress.discoverURL(webClient)
	if err != nil {
		return registryAddress, err
	}

	return registryAddress, nil
}

func (r registryAddress) String() string {
	switch r.registryType {
	case TypeProvider:
		return r.provider.String()
	case TypeModule:
		return r.module.String()
	default:
		logrus.Debugf("unknown registry type %q", r.registryType)
	}

	return ""
}

func (r registryAddress) ForDisplay() string {

	switch r.registryType {
	case TypeProvider:
		return r.provider.ForDisplay()
	case TypeModule:
		return r.module.ForDisplay()
	default:
		logrus.Debugf("unknown registry type %q", r.registryType)
	}

	return ""
}

func (r registryAddress) Hostname() string {
	switch r.registryType {
	case TypeProvider:
		return r.provider.Hostname.String()
	case TypeModule:
		return r.module.Package.Host.String()
	}

	logrus.Debugf("unknown registry type %q", r.registryType)
	return ""
}

func (r registryAddress) Path() string {
	switch r.registryType {
	case TypeProvider:
		return fmt.Sprintf("%s%s/%s/versions", r.wellKnown.ProviderPath, r.provider.Namespace, r.provider.Type)
	case TypeModule:
		return fmt.Sprintf("%s%s/versions", r.wellKnown.ModulesPath, r.module.Package.ForRegistryProtocol())
	}

	logrus.Debugf("unknown registry type %q", r.registryType)
	return ""
}

func (r registryAddress) API() string {
	return fmt.Sprintf("https://%s%s", r.Hostname(), r.Path())
}

func (r *registryAddress) discoverURL(webClient httpclient.HTTPClient) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/.well-known/terraform.json", r.Hostname()), nil)
	if err != nil {
		return err
	}

	res, err := webClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Debugf("\n%v\n", string(body))
		return err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("something went wrong while getting npm api data%q\n", err)
		return err
	}

	wellKnownResponse := wellKnownResponse{}

	err = json.Unmarshal(data, &wellKnownResponse)
	if err != nil {
		logrus.Errorf("error unmarshalling json: %q", err)
		return err
	}

	r.wellKnown = wellKnownResponse

	return nil
}
