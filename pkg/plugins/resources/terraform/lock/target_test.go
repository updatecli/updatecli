package lock

import (
	"errors"
	"testing"

	"github.com/minamijoyo/tfupdate/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestTarget(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		sourceInput      string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - No change",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			sourceInput:    "2.22.0",
			expectedResult: false,
		},
		{
			name: "Success - Expected change",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			sourceInput:    "2.23.0",
			expectedResult: true,
		},
		{
			name: "Success - Expected change using Value",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
				Value:     "2.23.0",
			},
			expectedResult: true,
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File:      "testdata/doNotExist.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ The specified file "testdata/doNotExist.hcl" does not exist`),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/null",
				Platforms: []string{"linux_amd64"},
				Value:     "3.2.1",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ cannot find value for "registry.terraform.io/hashicorp/null" from file "testdata/terraform.lock.hcl"`),
		},
		{
			name: "Failure - HTTP Target",
			spec: Spec{
				File:      "http://localhost/doNotExist.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ URL scheme is not supported for HCL target: "http://localhost/doNotExist.hcl"`),
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
		lock.NewMockProviderVersion(
			"hashicorp/null",
			"3.2.1",
			[]string{"linux_amd64"},
			map[string]string{
				"terraform-provider-null_3.2.1_linux_amd64.zip": "h1:FbGfc+muBsC17Ohy5g806iuI1hQc4SIexpYCrQHQd8w=",
			},
			map[string]string{
				"terraform-provider-null_3.2.1_freebsd_arm.zip":   "zh:58ed64389620cc7b82f01332e27723856422820cfd302e304b5f6c3436fb9840",
				"terraform-provider-null_3.2.1_manifest.json":     "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
				"terraform-provider-null_3.2.1_windows_amd64.zip": "zh:79e553aff77f1cfa9012a2218b8238dd672ea5e1b2924775ac9ac24d2a75c238",
				"terraform-provider-null_3.2.1_freebsd_amd64.zip": "zh:a1e06ddda0b5ac48f7e7c7d59e1ab5a4073bbcf876c73c0299e4610ed53859dc",
				"terraform-provider-null_3.2.1_linux_386.zip":     "zh:e80a746921946d8b6761e77305b752ad188da60688cfd2059322875d363be5f5",
				"terraform-provider-null_3.2.1_freebsd_386.zip":   "zh:fbdb892d9822ed0e4cb60f2fedbdbb556e4da0d88d3b942ae963ed6ff091e48f",
				"terraform-provider-null_3.2.1_windows_386.zip":   "zh:62a5cc82c3b2ddef7ef3a6f2fedb7b9b3deff4ab7b414938b08e51d6e8be87cb",
				"terraform-provider-null_3.2.1_darwin_amd64.zip":  "zh:63cff4de03af983175a7e37e52d4bd89d990be256b16b5c7f919aff5ad485aa5",
				"terraform-provider-null_3.2.1_linux_amd64.zip":   "zh:74cb22c6700e48486b7cabefa10b33b801dfcab56f1a6ac9b6624531f3d36ea3",
				"terraform-provider-null_3.2.1_linux_arm.zip":     "zh:c37a97090f1a82222925d45d84483b2aa702ef7ab66532af6cbcfb567818b970",
				"terraform-provider-null_3.2.1_darwin_arm64.zip":  "zh:e4453fbebf90c53ca3323a92e7ca0f9961427d2f0ce0d2b65523cc04d5d999c2",
				"terraform-provider-null_3.2.1_linux_arm64.zip":   "zh:fca01a623d90d0cad0843102f9b8b9fe0d3ff8244593bd817f126582b52dd694",
			},
		),
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.spec)

			require.NoError(t, err)

			l.lockIndex = lock.NewMockIndex(providerVersions)

			gotResult := result.Target{}
			err = l.Target(result.SourceInformation{Value: tt.sourceInput}, nil, true, &gotResult)
			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}
}
