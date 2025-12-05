package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	extism "github.com/extism/go-sdk"
	"github.com/go-viper/mapstructure/v2"
	"github.com/tetratelabs/wazero"

	"github.com/sirupsen/logrus"
)

// Plugin represents an Updatecli plugin object
type Plugin struct {
	manifest extism.Manifest
	spec     Spec
	rootDir  string
	scmID    string
	actionID string
}

func New(spec interface{}, rootDir, scmID, actionID, path string) (Plugin, error) {

	var manifest extism.Manifest
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Plugin{}, err
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		manifest = extism.Manifest{
			Wasm: []extism.Wasm{
				extism.WasmUrl{
					Url: path,
				},
			},
		}
	} else {
		manifest = extism.Manifest{
			Wasm: []extism.Wasm{
				extism.WasmFile{
					Path: path,
				},
			},
		}
	}

	logrus.Debugf("Configuring plugin allowed hosts: %v\n", s.AllowHosts)
	manifest.AllowedHosts = append(manifest.AllowedHosts, s.AllowHosts...)

	if s.AllowedPaths == nil {
		s.AllowedPaths = &[]string{
			".:/mnt",
		}
	}

	for k, v := range getAllowedPaths(rootDir, *s.AllowedPaths) {
		logrus.Debugf("Adding allowed path %q -> %q to plugin manifest\n", k, v)
		if manifest.AllowedPaths == nil {
			manifest.AllowedPaths = map[string]string{}
		}
		manifest.AllowedPaths[k] = v
	}

	return Plugin{
		manifest: manifest,
		spec:     s,
		rootDir:  rootDir,
		scmID:    scmID,
		actionID: actionID,
	}, nil
}

func (p Plugin) DiscoverManifests() ([][]byte, error) {

	ctx := context.Background()

	var timeout uint64 = 300 // seconds

	if p.spec.Timeout != nil {
		timeout = *p.spec.Timeout
	}

	pluginConfig := extism.PluginConfig{
		ModuleConfig:              wazero.NewModuleConfig().WithSysWalltime(),
		RuntimeConfig:             wazero.NewRuntimeConfig().WithCloseOnContextDone(timeout > 0),
		EnableWasi:                true,
		EnableHttpResponseHeaders: true,
	}

	if timeout > 0 {
		logrus.Debugf("Setting plugin timeout to %d seconds\n", timeout)
		p.manifest.Timeout = timeout
	}

	plugin, err := extism.NewPlugin(ctx, p.manifest, pluginConfig, []extism.HostFunction{
		extism.NewHostFunctionWithStack(
			"generate_docker_source_spec",
			generate_docker_source_spec,
			[]extism.ValueType{extism.ValueTypePTR},
			[]extism.ValueType{extism.ValueTypePTR},
		),
		extism.NewHostFunctionWithStack(
			"versionfilter_greater_than_pattern",
			versionfilter_greater_than_pattern,
			[]extism.ValueType{extism.ValueTypePTR},
			[]extism.ValueType{extism.ValueTypePTR},
		),
	})

	if err != nil {
		return nil, fmt.Errorf("creating plugin: %w", err)
	}

	defer plugin.Close(ctx)

	inputData := struct {
		ScmID    string         `json:"scmid"`
		ActionID string         `json:"actionid"`
		RootDir  string         `json:"rootdir"`
		Spec     map[string]any `json:"spec"`
	}{
		ScmID:    p.scmID,
		ActionID: p.actionID,
		RootDir:  p.rootDir,
		Spec:     p.spec.Spec,
	}

	input, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("formatting plugin input data: %w", err)
	}

	_, out, err := plugin.Call("_start", input)
	if err != nil {
		return nil, fmt.Errorf("calling plugin: %w", err)
	}

	type pluginOutput struct {
		Manifests []string `json:"manifests"`
	}

	var autodiscoveryOutput pluginOutput
	err = json.Unmarshal(out, &autodiscoveryOutput)
	if err != nil {
		return nil, fmt.Errorf("unable to parse autodiscovery plugin output: %w", err)
	}

	var manifests [][]byte
	for i := range autodiscoveryOutput.Manifests {
		manifests = append(manifests, []byte(autodiscoveryOutput.Manifests[i]))
	}

	return manifests, nil
}
