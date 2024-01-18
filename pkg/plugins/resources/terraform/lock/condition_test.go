package lock

import (
	"errors"
	"testing"

	"github.com/minamijoyo/tfupdate/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		source           string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - Using Value",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
				Value:     "2.22.0",
			},
			source:         "",
			expectedResult: true,
		},
		{
			name: "Success - Using Source",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			source:         "2.22.0",
			expectedResult: true,
		},
		{
			name: "Failure - Using Source",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			source:         "2.23.0",
			expectedResult: false,
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File:      "testdata/doNotExist.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			source:           "",
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ The specified file "testdata/doNotExist.hcl" does not exist`),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/null",
				Platforms: []string{"linux_amd64"},
			},
			source:           "",
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ cannot find value for "registry.terraform.io/hashicorp/null" from file "testdata/terraform.lock.hcl"`),
		},
		{
			name: "Failure - Multiple Files",
			spec: Spec{
				Files:     []string{"testdata/data.hcl", "testdata/data2.hcl"},
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ terraform/lock condition only supports one file"),
		},
	}

	providerVersions := []*lock.ProviderVersion{
		lock.NewMockProviderVersion(
			"hashicorp/kubernetes",
			"2.22.0",
			[]string{"linux_amd64"},
			map[string]string{
				"terraform-provider-kubernetes_2.22.0_linux_amd64.zip": "h1:b6Wj111/wsMNg8FrHFXrf4mCZFtSXKHx4JvbZh3YTCY=",
			},
			map[string]string{
				"terraform-provider-kubernetes_2.22.0_darwin_amd64.zip":  "zh:4c35651603493437b0b13e070148a330c034ac62c8967c2de9da6620b26adca4",
				"terraform-provider-kubernetes_2.22.0_windows_386.zip":   "zh:6fc0c2d65864324edde4db84f528268065df58229fc3ee321626687b0e603637",
				"terraform-provider-kubernetes_2.22.0_windows_amd64.zip": "zh:73c58d007aba7f67c0aa9029794e10c2517bec565b7cb57d0f5948ea3f30e407",
				"terraform-provider-kubernetes_2.22.0_freebsd_amd64.zip": "zh:7d6fc9d3c1843baccd2e1fc56317925a2f9df372427d30fcb5052d123adc887a",
				"terraform-provider-kubernetes_2.22.0_manifest.json":     "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
				"terraform-provider-kubernetes_2.22.0_linux_arm.zip":     "zh:1eac662b1f238042b2068401e510f0624efaf51fd6a4dd9c49d710a49d383b61",
				"terraform-provider-kubernetes_2.22.0_linux_amd64.zip":   "zh:50c0e8654efb46e3a3666c638ca2e0c8aec07f985fbc80f9205bed960386dc9b",
				"terraform-provider-kubernetes_2.22.0_darwin_arm64.zip":  "zh:5f65194ddd6ea7e89b378297d882083a4b84962edb35dd35752f0c7e9d6282a0",
				"terraform-provider-kubernetes_2.22.0_linux_386.zip":     "zh:a0ad9eb863b51586ea306c5f2beef74476c96684aed41a3ee99eb4b6d8898d01",
				"terraform-provider-kubernetes_2.22.0_freebsd_386.zip":   "zh:e218fcfbf4994ff741408a023a9d9eb6c697ce9f63ce5540d3b35226d86c963e",
				"terraform-provider-kubernetes_2.22.0_linux_arm64.zip":   "zh:f95625f317795f0e38cc6293dd31c85863f4e225209d07d1e233c50d9295083c",
				"terraform-provider-kubernetes_2.22.0_freebsd_arm.zip":   "zh:f96e0923a632bc430267fe915794972be873887f5e761ed11451d67202e256c8",
			},
		),
		lock.NewMockProviderVersion(
			"hashicorp/kubernetes",
			"2.23.0",
			[]string{"linux_amd64"},
			map[string]string{
				"terraform-provider-kubernetes_2.23.0_linux_amd64.zip": "h1:xyFc77aYkPoU4Xt1i5t0B1IaS8TbTtp9aCSuQKDayII=",
			},
			map[string]string{
				"terraform-provider-kubernetes_2.23.0_freebsd_386.zip":   "zh:1102ba5ca1a595f880e67102bbf999cc8b60203272a078a5b1e896d173f3f34b",
				"terraform-provider-kubernetes_2.23.0_linux_386.zip":     "zh:1347cf958ed3f3f80b3c7b3e23ddda3d6c6573a81847a8ee92b7df231c238bf6",
				"terraform-provider-kubernetes_2.23.0_linux_arm64.zip":   "zh:2cb18e9f5156bc1b1ee6bc580a709f7c2737d142722948f4a6c3c8efe757fa8d",
				"terraform-provider-kubernetes_2.23.0_darwin_amd64.zip":  "zh:5506aa6f28dcca2a265ccf8e34478b5ec2cb43b867fe6d93b0158f01590fdadd",
				"terraform-provider-kubernetes_2.23.0_darwin_arm64.zip":  "zh:6217a20686b631b1dcb448ee4bc795747ebc61b56fbe97a1ad51f375ebb0d996",
				"terraform-provider-kubernetes_2.23.0_linux_arm.zip":     "zh:e6068e3f94f6943b5586557b73f109debe19d1a75ca9273a681d22d1ce066579",
				"terraform-provider-kubernetes_2.23.0_windows_386.zip":   "zh:10488a12525ed674359585f83e3ee5e74818b5c98e033798351678b21b2f7d89",
				"terraform-provider-kubernetes_2.23.0_freebsd_arm.zip":   "zh:8accf916c00579c22806cb771e8909b349ffb7eb29d9c5468d0a3f3166c7a84a",
				"terraform-provider-kubernetes_2.23.0_freebsd_amd64.zip": "zh:9379b0b54a0fa030b19c7b9356708ec8489e194c3b5e978df2d31368563308e5",
				"terraform-provider-kubernetes_2.23.0_windows_amd64.zip": "zh:aa99c580890691036c2931841e88e7ee80d59ae52289c8c2c28ea0ac23e31520",
				"terraform-provider-kubernetes_2.23.0_linux_amd64.zip":   "zh:c57376d169875990ac68664d227fb69cd0037b92d0eba6921d757c3fd1879080",
				"terraform-provider-kubernetes_2.23.0_manifest.json":     "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
			},
		),
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.spec)

			require.NoError(t, err)

			l.lockIndex = lock.NewMockIndex(providerVersions)

			gotResult, _, gotErr := l.Condition(tt.source, nil)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), gotErr.Error())
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
