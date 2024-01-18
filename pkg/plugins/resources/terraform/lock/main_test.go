package lock

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	expectedHashes := []string{
		"h1:b6Wj111/wsMNg8FrHFXrf4mCZFtSXKHx4JvbZh3YTCY=",
		"zh:1eac662b1f238042b2068401e510f0624efaf51fd6a4dd9c49d710a49d383b61",
		"zh:4c35651603493437b0b13e070148a330c034ac62c8967c2de9da6620b26adca4",
		"zh:50c0e8654efb46e3a3666c638ca2e0c8aec07f985fbc80f9205bed960386dc9b",
		"zh:5f65194ddd6ea7e89b378297d882083a4b84962edb35dd35752f0c7e9d6282a0",
		"zh:6fc0c2d65864324edde4db84f528268065df58229fc3ee321626687b0e603637",
		"zh:73c58d007aba7f67c0aa9029794e10c2517bec565b7cb57d0f5948ea3f30e407",
		"zh:7d6fc9d3c1843baccd2e1fc56317925a2f9df372427d30fcb5052d123adc887a",
		"zh:a0ad9eb863b51586ea306c5f2beef74476c96684aed41a3ee99eb4b6d8898d01",
		"zh:e218fcfbf4994ff741408a023a9d9eb6c697ce9f63ce5540d3b35226d86c963e",
		"zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
		"zh:f95625f317795f0e38cc6293dd31c85863f4e225209d07d1e233c50d9295083c",
		"zh:f96e0923a632bc430267fe915794972be873887f5e761ed11451d67202e256c8",
	}

	testData := []struct {
		name                  string
		spec                  Spec
		workingDir            string
		expectedErrorMsg      error
		wantErr               bool
		expectedResultVersion string
		expectedResultHashes  []string
	}{
		{
			name: "Success - Query file",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			wantErr:               false,
			workingDir:            "",
			expectedResultVersion: `2.22.0`,
			expectedResultHashes:  expectedHashes,
		},
		{
			name: "Success - Query files",
			spec: Spec{
				Files:     []string{"testdata/terraform.lock.hcl"},
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			wantErr:               false,
			workingDir:            "",
			expectedResultVersion: `2.22.0`,
			expectedResultHashes:  expectedHashes,
		},
		{
			name: "Success - Query file, short name",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			wantErr:               false,
			workingDir:            "",
			expectedResultVersion: `2.22.0`,
			expectedResultHashes:  expectedHashes,
		},
		{
			name: "Failure - missing",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "hashicorp/null",
				Platforms: []string{"linux_amd64"},
			},
			wantErr:          true,
			workingDir:       "",
			expectedErrorMsg: errors.New(`✗ cannot find value for "registry.terraform.io/hashicorp/null" from file "testdata/terraform.lock.hcl"`),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			err = h.Read()

			require.NoError(t, err)

			resourceFile := h.files["testdata/terraform.lock.hcl"]

			resultVersion, resultHashes, err := h.Query(resourceFile)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResultVersion, resultVersion)
			assert.Equal(t, tt.expectedResultHashes, resultHashes)
		})
	}
}

func TestApply(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		value            string
		hashes           []string
		workingDir       string
		expectedErrorMsg error
		wantErr          bool
		expectedResult   string
	}{
		{
			name: "Success - Update FQ",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			value: "2.23.0",
			hashes: []string{
				"h1:xyFc77aYkPoU4Xt1i5t0B1IaS8TbTtp9aCSuQKDayII=",
				"zh:10488a12525ed674359585f83e3ee5e74818b5c98e033798351678b21b2f7d89",
				"zh:1102ba5ca1a595f880e67102bbf999cc8b60203272a078a5b1e896d173f3f34b",
				"zh:1347cf958ed3f3f80b3c7b3e23ddda3d6c6573a81847a8ee92b7df231c238bf6",
				"zh:2cb18e9f5156bc1b1ee6bc580a709f7c2737d142722948f4a6c3c8efe757fa8d",
				"zh:5506aa6f28dcca2a265ccf8e34478b5ec2cb43b867fe6d93b0158f01590fdadd",
				"zh:6217a20686b631b1dcb448ee4bc795747ebc61b56fbe97a1ad51f375ebb0d996",
				"zh:8accf916c00579c22806cb771e8909b349ffb7eb29d9c5468d0a3f3166c7a84a",
				"zh:9379b0b54a0fa030b19c7b9356708ec8489e194c3b5e978df2d31368563308e5",
				"zh:aa99c580890691036c2931841e88e7ee80d59ae52289c8c2c28ea0ac23e31520",
				"zh:c57376d169875990ac68664d227fb69cd0037b92d0eba6921d757c3fd1879080",
				"zh:e6068e3f94f6943b5586557b73f109debe19d1a75ca9273a681d22d1ce066579",
				"zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
			},
			wantErr:    false,
			workingDir: "",
			expectedResult: `# This file is maintained automatically by "terraform init".
# Manual edits may be lost in future updates.

provider "registry.terraform.io/hashicorp/aws" {
  version = "5.9.0"
  hashes = [
    "h1:mvg6WWqqUvgUq6wYCWg/zqpND/5yIz3plIL1IOR50Rs=",
    "zh:032424d4686ce2ff7c5a4a738491635616afbf6e06b3e7e6a754baa031d1265d",
    "zh:1e530b4020544ec94e1fe7b1e4296640eb12cf1bf4f79cd6429ff2c4e6fffaf3",
    "zh:24d2eee57a4c78039959dd9bb6dff2b75ed0483d44929550c067c3488307dc62",
    "zh:3ad6d736722059664e790a358eacf0e0e60973ec44e70142fb503275de2116c1",
    "zh:3f34d81acf86c61ddd271e9c4b8215765037463c3fe3c7aea1dc32a509020cfb",
    "zh:65a04aa615fc320059a0871702c83b6be10bce2064056096b46faffe768a698e",
    "zh:7fb56c3ce1fe77983627e2931e7c7b73152180c4dfb03e793413d0137c85d6b2",
    "zh:90c94cb9d7352468bcd5ba21a56099fe087a072b1936d86f47d54c2a012b708a",
    "zh:9b12af85486a96aedd8d7984b0ff811a4b42e3d88dad1a3fb4c0b580d04fa425",
    "zh:a109c5f01ed48852fe17847fa8a116dfdb81500794a9cf7e5ef92ea6dec20431",
    "zh:a27c5396077a36ac2801d4c1c1132201a9225a65bba0e3b3aded9cc18f2c38ff",
    "zh:a86ad796ccb0f2cb8f0ca069c774dbf74964edd3282529726816c72e22164b3c",
    "zh:bda8afc64091a2a72e0cc38fde937b2163b1b072a5c41310d255901207571afd",
    "zh:d22473894cd7e94b7a971793dd07309569f82913a10e4bd6c22e04f362f03bb9",
    "zh:f4dbb6d13511290a5274f5b202e6d9997643f86e4c48e8c5e3c204121082851a",
  ]
}

provider "registry.terraform.io/hashicorp/cloudinit" {
  version = "2.3.2"
  hashes = [
    "h1:Vl0aixAYTV/bjathX7VArC5TVNkxBCsi3Vq7R4z1uvc=",
    "zh:2487e498736ed90f53de8f66fe2b8c05665b9f8ff1506f751c5ee227c7f457d1",
    "zh:3d8627d142942336cf65eea6eb6403692f47e9072ff3fa11c3f774a3b93130b3",
    "zh:434b643054aeafb5df28d5529b72acc20c6f5ded24decad73b98657af2b53f4f",
    "zh:436aa6c2b07d82aa6a9dd746a3e3a627f72787c27c80552ceda6dc52d01f4b6f",
    "zh:458274c5aabe65ef4dbd61d43ce759287788e35a2da004e796373f88edcaa422",
    "zh:54bc70fa6fb7da33292ae4d9ceef5398d637c7373e729ed4fce59bd7b8d67372",
    "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
    "zh:893ba267e18749c1a956b69be569f0d7bc043a49c3a0eb4d0d09a8e8b2ca3136",
    "zh:95493b7517bce116f75cdd4c63b7c82a9d0d48ec2ef2f5eb836d262ef96d0aa7",
    "zh:9ae21ab393be52e3e84e5cce0ef20e690d21f6c10ade7d9d9d22b39851bfeddc",
    "zh:cc3b01ac2472e6d59358d54d5e4945032efbc8008739a6d4946ca1b621a16040",
    "zh:f23bfe9758f06a1ec10ea3a81c9deedf3a7b42963568997d84a5153f35c5839a",
  ]
}

provider "registry.terraform.io/hashicorp/kubernetes" {
  version = "2.23.0"
  hashes = [
    "h1:xyFc77aYkPoU4Xt1i5t0B1IaS8TbTtp9aCSuQKDayII=",
    "zh:10488a12525ed674359585f83e3ee5e74818b5c98e033798351678b21b2f7d89",
    "zh:1102ba5ca1a595f880e67102bbf999cc8b60203272a078a5b1e896d173f3f34b",
    "zh:1347cf958ed3f3f80b3c7b3e23ddda3d6c6573a81847a8ee92b7df231c238bf6",
    "zh:2cb18e9f5156bc1b1ee6bc580a709f7c2737d142722948f4a6c3c8efe757fa8d",
    "zh:5506aa6f28dcca2a265ccf8e34478b5ec2cb43b867fe6d93b0158f01590fdadd",
    "zh:6217a20686b631b1dcb448ee4bc795747ebc61b56fbe97a1ad51f375ebb0d996",
    "zh:8accf916c00579c22806cb771e8909b349ffb7eb29d9c5468d0a3f3166c7a84a",
    "zh:9379b0b54a0fa030b19c7b9356708ec8489e194c3b5e978df2d31368563308e5",
    "zh:aa99c580890691036c2931841e88e7ee80d59ae52289c8c2c28ea0ac23e31520",
    "zh:c57376d169875990ac68664d227fb69cd0037b92d0eba6921d757c3fd1879080",
    "zh:e6068e3f94f6943b5586557b73f109debe19d1a75ca9273a681d22d1ce066579",
    "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
  ]
}

provider "registry.terraform.io/hashicorp/local" {
  version = "2.4.0"
  hashes = [
    "h1:R97FTYETo88sT2VHfMgkPU3lzCsZLunPftjSI5vfKe8=",
    "zh:53604cd29cb92538668fe09565c739358dc53ca56f9f11312b9d7de81e48fab9",
    "zh:66a46e9c508716a1c98efbf793092f03d50049fa4a83cd6b2251e9a06aca2acf",
    "zh:70a6f6a852dd83768d0778ce9817d81d4b3f073fab8fa570bff92dcb0824f732",
    "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
    "zh:82a803f2f484c8b766e2e9c32343e9c89b91997b9f8d2697f9f3837f62926b35",
    "zh:9708a4e40d6cc4b8afd1352e5186e6e1502f6ae599867c120967aebe9d90ed04",
    "zh:973f65ce0d67c585f4ec250c1e634c9b22d9c4288b484ee2a871d7fa1e317406",
    "zh:c8fa0f98f9316e4cfef082aa9b785ba16e36ff754d6aba8b456dab9500e671c6",
    "zh:cfa5342a5f5188b20db246c73ac823918c189468e1382cb3c48a9c0c08fc5bf7",
    "zh:e0e2b477c7e899c63b06b38cd8684a893d834d6d0b5e9b033cedc06dd7ffe9e2",
    "zh:f62d7d05ea1ee566f732505200ab38d94315a4add27947a60afa29860822d3fc",
    "zh:fa7ce69dde358e172bd719014ad637634bbdabc49363104f4fca759b4b73f2ce",
  ]
}

provider "registry.terraform.io/hashicorp/random" {
  version = "3.5.1"
  hashes = [
    "h1:VSnd9ZIPyfKHOObuQCaKfnjIHRtR7qTw19Rz8tJxm+k=",
    "zh:04e3fbd610cb52c1017d282531364b9c53ef72b6bc533acb2a90671957324a64",
    "zh:119197103301ebaf7efb91df8f0b6e0dd31e6ff943d231af35ee1831c599188d",
    "zh:4d2b219d09abf3b1bb4df93d399ed156cadd61f44ad3baf5cf2954df2fba0831",
    "zh:6130bdde527587bbe2dcaa7150363e96dbc5250ea20154176d82bc69df5d4ce3",
    "zh:6cc326cd4000f724d3086ee05587e7710f032f94fc9af35e96a386a1c6f2214f",
    "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
    "zh:b6d88e1d28cf2dfa24e9fdcc3efc77adcdc1c3c3b5c7ce503a423efbdd6de57b",
    "zh:ba74c592622ecbcef9dc2a4d81ed321c4e44cddf7da799faa324da9bf52a22b2",
    "zh:c7c5cde98fe4ef1143bd1b3ec5dc04baf0d4cc3ca2c5c7d40d17c0e9b2076865",
    "zh:dac4bad52c940cd0dfc27893507c1e92393846b024c5a9db159a93c534a3da03",
    "zh:de8febe2a2acd9ac454b844a4106ed295ae9520ef54dc8ed2faf29f12716b602",
    "zh:eab0d0495e7e711cca367f7d4df6e322e6c562fc52151ec931176115b83ed014",
  ]
}

provider "registry.terraform.io/hashicorp/tls" {
  version = "4.0.4"
  hashes = [
    "h1:pe9vq86dZZKCm+8k1RhzARwENslF3SXb9ErHbQfgjXU=",
    "zh:23671ed83e1fcf79745534841e10291bbf34046b27d6e68a5d0aab77206f4a55",
    "zh:45292421211ffd9e8e3eb3655677700e3c5047f71d8f7650d2ce30242335f848",
    "zh:59fedb519f4433c0fdb1d58b27c210b27415fddd0cd73c5312530b4309c088be",
    "zh:5a8eec2409a9ff7cd0758a9d818c74bcba92a240e6c5e54b99df68fff312bbd5",
    "zh:5e6a4b39f3171f53292ab88058a59e64825f2b842760a4869e64dc1dc093d1fe",
    "zh:810547d0bf9311d21c81cc306126d3547e7bd3f194fc295836acf164b9f8424e",
    "zh:824a5f3617624243bed0259d7dd37d76017097dc3193dac669be342b90b2ab48",
    "zh:9361ccc7048be5dcbc2fafe2d8216939765b3160bd52734f7a9fd917a39ecbd8",
    "zh:aa02ea625aaf672e649296bce7580f62d724268189fe9ad7c1b36bb0fa12fa60",
    "zh:c71b4cd40d6ec7815dfeefd57d88bc592c0c42f5e5858dcc88245d371b4b8b1e",
    "zh:dabcd52f36b43d250a3d71ad7abfa07b5622c69068d989e60b79b2bb4f220316",
    "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
  ]
}
`,
		},
		{
			name: "Success - Update short name",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			value: "2.23.0",
			hashes: []string{
				"h1:xyFc77aYkPoU4Xt1i5t0B1IaS8TbTtp9aCSuQKDayII=",
				"zh:10488a12525ed674359585f83e3ee5e74818b5c98e033798351678b21b2f7d89",
				"zh:1102ba5ca1a595f880e67102bbf999cc8b60203272a078a5b1e896d173f3f34b",
				"zh:1347cf958ed3f3f80b3c7b3e23ddda3d6c6573a81847a8ee92b7df231c238bf6",
				"zh:2cb18e9f5156bc1b1ee6bc580a709f7c2737d142722948f4a6c3c8efe757fa8d",
				"zh:5506aa6f28dcca2a265ccf8e34478b5ec2cb43b867fe6d93b0158f01590fdadd",
				"zh:6217a20686b631b1dcb448ee4bc795747ebc61b56fbe97a1ad51f375ebb0d996",
				"zh:8accf916c00579c22806cb771e8909b349ffb7eb29d9c5468d0a3f3166c7a84a",
				"zh:9379b0b54a0fa030b19c7b9356708ec8489e194c3b5e978df2d31368563308e5",
				"zh:aa99c580890691036c2931841e88e7ee80d59ae52289c8c2c28ea0ac23e31520",
				"zh:c57376d169875990ac68664d227fb69cd0037b92d0eba6921d757c3fd1879080",
				"zh:e6068e3f94f6943b5586557b73f109debe19d1a75ca9273a681d22d1ce066579",
				"zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
			},
			wantErr:    false,
			workingDir: "",
			expectedResult: `# This file is maintained automatically by "terraform init".
# Manual edits may be lost in future updates.

provider "registry.terraform.io/hashicorp/aws" {
  version = "5.9.0"
  hashes = [
    "h1:mvg6WWqqUvgUq6wYCWg/zqpND/5yIz3plIL1IOR50Rs=",
    "zh:032424d4686ce2ff7c5a4a738491635616afbf6e06b3e7e6a754baa031d1265d",
    "zh:1e530b4020544ec94e1fe7b1e4296640eb12cf1bf4f79cd6429ff2c4e6fffaf3",
    "zh:24d2eee57a4c78039959dd9bb6dff2b75ed0483d44929550c067c3488307dc62",
    "zh:3ad6d736722059664e790a358eacf0e0e60973ec44e70142fb503275de2116c1",
    "zh:3f34d81acf86c61ddd271e9c4b8215765037463c3fe3c7aea1dc32a509020cfb",
    "zh:65a04aa615fc320059a0871702c83b6be10bce2064056096b46faffe768a698e",
    "zh:7fb56c3ce1fe77983627e2931e7c7b73152180c4dfb03e793413d0137c85d6b2",
    "zh:90c94cb9d7352468bcd5ba21a56099fe087a072b1936d86f47d54c2a012b708a",
    "zh:9b12af85486a96aedd8d7984b0ff811a4b42e3d88dad1a3fb4c0b580d04fa425",
    "zh:a109c5f01ed48852fe17847fa8a116dfdb81500794a9cf7e5ef92ea6dec20431",
    "zh:a27c5396077a36ac2801d4c1c1132201a9225a65bba0e3b3aded9cc18f2c38ff",
    "zh:a86ad796ccb0f2cb8f0ca069c774dbf74964edd3282529726816c72e22164b3c",
    "zh:bda8afc64091a2a72e0cc38fde937b2163b1b072a5c41310d255901207571afd",
    "zh:d22473894cd7e94b7a971793dd07309569f82913a10e4bd6c22e04f362f03bb9",
    "zh:f4dbb6d13511290a5274f5b202e6d9997643f86e4c48e8c5e3c204121082851a",
  ]
}

provider "registry.terraform.io/hashicorp/cloudinit" {
  version = "2.3.2"
  hashes = [
    "h1:Vl0aixAYTV/bjathX7VArC5TVNkxBCsi3Vq7R4z1uvc=",
    "zh:2487e498736ed90f53de8f66fe2b8c05665b9f8ff1506f751c5ee227c7f457d1",
    "zh:3d8627d142942336cf65eea6eb6403692f47e9072ff3fa11c3f774a3b93130b3",
    "zh:434b643054aeafb5df28d5529b72acc20c6f5ded24decad73b98657af2b53f4f",
    "zh:436aa6c2b07d82aa6a9dd746a3e3a627f72787c27c80552ceda6dc52d01f4b6f",
    "zh:458274c5aabe65ef4dbd61d43ce759287788e35a2da004e796373f88edcaa422",
    "zh:54bc70fa6fb7da33292ae4d9ceef5398d637c7373e729ed4fce59bd7b8d67372",
    "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
    "zh:893ba267e18749c1a956b69be569f0d7bc043a49c3a0eb4d0d09a8e8b2ca3136",
    "zh:95493b7517bce116f75cdd4c63b7c82a9d0d48ec2ef2f5eb836d262ef96d0aa7",
    "zh:9ae21ab393be52e3e84e5cce0ef20e690d21f6c10ade7d9d9d22b39851bfeddc",
    "zh:cc3b01ac2472e6d59358d54d5e4945032efbc8008739a6d4946ca1b621a16040",
    "zh:f23bfe9758f06a1ec10ea3a81c9deedf3a7b42963568997d84a5153f35c5839a",
  ]
}

provider "registry.terraform.io/hashicorp/kubernetes" {
  version = "2.23.0"
  hashes = [
    "h1:xyFc77aYkPoU4Xt1i5t0B1IaS8TbTtp9aCSuQKDayII=",
    "zh:10488a12525ed674359585f83e3ee5e74818b5c98e033798351678b21b2f7d89",
    "zh:1102ba5ca1a595f880e67102bbf999cc8b60203272a078a5b1e896d173f3f34b",
    "zh:1347cf958ed3f3f80b3c7b3e23ddda3d6c6573a81847a8ee92b7df231c238bf6",
    "zh:2cb18e9f5156bc1b1ee6bc580a709f7c2737d142722948f4a6c3c8efe757fa8d",
    "zh:5506aa6f28dcca2a265ccf8e34478b5ec2cb43b867fe6d93b0158f01590fdadd",
    "zh:6217a20686b631b1dcb448ee4bc795747ebc61b56fbe97a1ad51f375ebb0d996",
    "zh:8accf916c00579c22806cb771e8909b349ffb7eb29d9c5468d0a3f3166c7a84a",
    "zh:9379b0b54a0fa030b19c7b9356708ec8489e194c3b5e978df2d31368563308e5",
    "zh:aa99c580890691036c2931841e88e7ee80d59ae52289c8c2c28ea0ac23e31520",
    "zh:c57376d169875990ac68664d227fb69cd0037b92d0eba6921d757c3fd1879080",
    "zh:e6068e3f94f6943b5586557b73f109debe19d1a75ca9273a681d22d1ce066579",
    "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
  ]
}

provider "registry.terraform.io/hashicorp/local" {
  version = "2.4.0"
  hashes = [
    "h1:R97FTYETo88sT2VHfMgkPU3lzCsZLunPftjSI5vfKe8=",
    "zh:53604cd29cb92538668fe09565c739358dc53ca56f9f11312b9d7de81e48fab9",
    "zh:66a46e9c508716a1c98efbf793092f03d50049fa4a83cd6b2251e9a06aca2acf",
    "zh:70a6f6a852dd83768d0778ce9817d81d4b3f073fab8fa570bff92dcb0824f732",
    "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
    "zh:82a803f2f484c8b766e2e9c32343e9c89b91997b9f8d2697f9f3837f62926b35",
    "zh:9708a4e40d6cc4b8afd1352e5186e6e1502f6ae599867c120967aebe9d90ed04",
    "zh:973f65ce0d67c585f4ec250c1e634c9b22d9c4288b484ee2a871d7fa1e317406",
    "zh:c8fa0f98f9316e4cfef082aa9b785ba16e36ff754d6aba8b456dab9500e671c6",
    "zh:cfa5342a5f5188b20db246c73ac823918c189468e1382cb3c48a9c0c08fc5bf7",
    "zh:e0e2b477c7e899c63b06b38cd8684a893d834d6d0b5e9b033cedc06dd7ffe9e2",
    "zh:f62d7d05ea1ee566f732505200ab38d94315a4add27947a60afa29860822d3fc",
    "zh:fa7ce69dde358e172bd719014ad637634bbdabc49363104f4fca759b4b73f2ce",
  ]
}

provider "registry.terraform.io/hashicorp/random" {
  version = "3.5.1"
  hashes = [
    "h1:VSnd9ZIPyfKHOObuQCaKfnjIHRtR7qTw19Rz8tJxm+k=",
    "zh:04e3fbd610cb52c1017d282531364b9c53ef72b6bc533acb2a90671957324a64",
    "zh:119197103301ebaf7efb91df8f0b6e0dd31e6ff943d231af35ee1831c599188d",
    "zh:4d2b219d09abf3b1bb4df93d399ed156cadd61f44ad3baf5cf2954df2fba0831",
    "zh:6130bdde527587bbe2dcaa7150363e96dbc5250ea20154176d82bc69df5d4ce3",
    "zh:6cc326cd4000f724d3086ee05587e7710f032f94fc9af35e96a386a1c6f2214f",
    "zh:78d5eefdd9e494defcb3c68d282b8f96630502cac21d1ea161f53cfe9bb483b3",
    "zh:b6d88e1d28cf2dfa24e9fdcc3efc77adcdc1c3c3b5c7ce503a423efbdd6de57b",
    "zh:ba74c592622ecbcef9dc2a4d81ed321c4e44cddf7da799faa324da9bf52a22b2",
    "zh:c7c5cde98fe4ef1143bd1b3ec5dc04baf0d4cc3ca2c5c7d40d17c0e9b2076865",
    "zh:dac4bad52c940cd0dfc27893507c1e92393846b024c5a9db159a93c534a3da03",
    "zh:de8febe2a2acd9ac454b844a4106ed295ae9520ef54dc8ed2faf29f12716b602",
    "zh:eab0d0495e7e711cca367f7d4df6e322e6c562fc52151ec931176115b83ed014",
  ]
}

provider "registry.terraform.io/hashicorp/tls" {
  version = "4.0.4"
  hashes = [
    "h1:pe9vq86dZZKCm+8k1RhzARwENslF3SXb9ErHbQfgjXU=",
    "zh:23671ed83e1fcf79745534841e10291bbf34046b27d6e68a5d0aab77206f4a55",
    "zh:45292421211ffd9e8e3eb3655677700e3c5047f71d8f7650d2ce30242335f848",
    "zh:59fedb519f4433c0fdb1d58b27c210b27415fddd0cd73c5312530b4309c088be",
    "zh:5a8eec2409a9ff7cd0758a9d818c74bcba92a240e6c5e54b99df68fff312bbd5",
    "zh:5e6a4b39f3171f53292ab88058a59e64825f2b842760a4869e64dc1dc093d1fe",
    "zh:810547d0bf9311d21c81cc306126d3547e7bd3f194fc295836acf164b9f8424e",
    "zh:824a5f3617624243bed0259d7dd37d76017097dc3193dac669be342b90b2ab48",
    "zh:9361ccc7048be5dcbc2fafe2d8216939765b3160bd52734f7a9fd917a39ecbd8",
    "zh:aa02ea625aaf672e649296bce7580f62d724268189fe9ad7c1b36bb0fa12fa60",
    "zh:c71b4cd40d6ec7815dfeefd57d88bc592c0c42f5e5858dcc88245d371b4b8b1e",
    "zh:dabcd52f36b43d250a3d71ad7abfa07b5622c69068d989e60b79b2bb4f220316",
    "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
  ]
}
`,
		},
		{
			name: "Success - Update with constraints",
			spec: Spec{
				File:      "testdata/terraform-constraints.lock.hcl",
				Provider:  "hashicorp/azurerm",
				Platforms: []string{"linux_amd64"},
			},
			value: "3.71.0",
			hashes: []string{
				"h1:QI0iaPNi0qAOIbXptd4ZObi0D5X1jojom5774GtEspA=",
				"zh:06f0d225b1711dfad256ff33134f878acc8f84624d9da66b075b075cc4d75892",
				"zh:09ff74056818babe02ea5a633bffe2b8223eaf79916dc1db169651ef7725c22f",
				"zh:27687e0f8458e6d88ebea94352eb523f56e8f5cdc468268af8f38dc4a4265bf4",
				"zh:2d81bfab3c6a9b897fa8fbb5256c9e5a944e6ecbf7f73a2a3e2b53a2c4fbcfc5",
				"zh:4cfc744cfc37aeeeecd82800c70e2591b38447af9e3c51bcbf06a5efe842ed65",
				"zh:734fbb81508b264f772a076338ddf1c7b25534d2007a1738a7d55587478ed258",
				"zh:9a5502c364f58073599fff8cdd8adc32e7f7bcd00a4d9b57d2fff678fd8a8319",
				"zh:9bc528f7e78dbfd106f94b741b68dedd3dd3d31c3defcddcc1972c8e52a6b7db",
				"zh:c30db03d877f9a7ae0c19d3fd338bbf95cdddbf6df1023709dbfa99689abac14",
				"zh:c51d4065145b8f4ca45fc9a0f3ca7f2d933bc0302af2eead74f3ce64a9221ae8",
				"zh:e23029fc7f81723795d7da770131adb1ce6f4d32f0a57eb75d47e036a0a19833",
				"zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
			},
			wantErr:    false,
			workingDir: "",
			expectedResult: `# This file is maintained automatically by "terraform init".
# Manual edits may be lost in future updates.

provider "registry.terraform.io/hashicorp/azuread" {
  version     = "2.41.0"
  constraints = "~> 2.39"
  hashes = [
    "h1:k3SUdcspQB5NtSux67Ngh7mkq/J3HW/6FllVhwK8vu8=",
    "zh:1c3e89cf19118fc07d7b04257251fc9897e722c16e0a0df7b07fcd261f8c12e7",
    "zh:32032e1539da9c986adea46a4fc702b8e3dea287ac825fdbc57ce459e6cfd25b",
    "zh:37841384780140f7026faf55700087d5e08f504168d67ae7f0d28e7132430c82",
    "zh:4c272ac7bf8b1aa913eeeca5fcdff25af432c6ae397d963710004b546a068936",
    "zh:8febf43db3f0bd9d46b2ee31eecf8f3ba912a4553ae049657e0f0c68bdff90a0",
    "zh:b60a3e92cd2a7af916d70d94d01c67d5bfabf68963f1b5ff75d8a310d66a4522",
    "zh:b9a8b9554f8ba9d306427aa8b2c0894242286972a61ab33300d6d8ac57bf93ab",
    "zh:ce7a79964a68a6086fa97e1f38d1c015d24d0732dd7876e179704c383a79f9aa",
    "zh:e305876a4d44739135264bb11be2c5489903682bb89f7b25561d38be31a9087d",
    "zh:ec3f63a848b3b2521cda0cf7650c4f85b50cc8905e5065f775d7c855833ab306",
    "zh:ef2a1ba97f15db88510bbc9b1af611c9bc8bee660d4d8b3b826424e1025487aa",
    "zh:f40db067d868567e199e16054a554aba5d89c3a19a4264c813dc7212724eeeea",
  ]
}

provider "registry.terraform.io/hashicorp/azurerm" {
  version     = "3.71.0"
  constraints = "3.71.0"
  hashes = [
    "h1:QI0iaPNi0qAOIbXptd4ZObi0D5X1jojom5774GtEspA=",
    "zh:06f0d225b1711dfad256ff33134f878acc8f84624d9da66b075b075cc4d75892",
    "zh:09ff74056818babe02ea5a633bffe2b8223eaf79916dc1db169651ef7725c22f",
    "zh:27687e0f8458e6d88ebea94352eb523f56e8f5cdc468268af8f38dc4a4265bf4",
    "zh:2d81bfab3c6a9b897fa8fbb5256c9e5a944e6ecbf7f73a2a3e2b53a2c4fbcfc5",
    "zh:4cfc744cfc37aeeeecd82800c70e2591b38447af9e3c51bcbf06a5efe842ed65",
    "zh:734fbb81508b264f772a076338ddf1c7b25534d2007a1738a7d55587478ed258",
    "zh:9a5502c364f58073599fff8cdd8adc32e7f7bcd00a4d9b57d2fff678fd8a8319",
    "zh:9bc528f7e78dbfd106f94b741b68dedd3dd3d31c3defcddcc1972c8e52a6b7db",
    "zh:c30db03d877f9a7ae0c19d3fd338bbf95cdddbf6df1023709dbfa99689abac14",
    "zh:c51d4065145b8f4ca45fc9a0f3ca7f2d933bc0302af2eead74f3ce64a9221ae8",
    "zh:e23029fc7f81723795d7da770131adb1ce6f4d32f0a57eb75d47e036a0a19833",
    "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
  ]
}
`,
		},

		{
			name: "Success - Update with skip constraints",
			spec: Spec{
				File:            "testdata/terraform-constraints.lock.hcl",
				Provider:        "hashicorp/azurerm",
				Platforms:       []string{"linux_amd64"},
				SkipConstraints: true,
			},
			value: "3.71.0",
			hashes: []string{
				"h1:QI0iaPNi0qAOIbXptd4ZObi0D5X1jojom5774GtEspA=",
				"zh:06f0d225b1711dfad256ff33134f878acc8f84624d9da66b075b075cc4d75892",
				"zh:09ff74056818babe02ea5a633bffe2b8223eaf79916dc1db169651ef7725c22f",
				"zh:27687e0f8458e6d88ebea94352eb523f56e8f5cdc468268af8f38dc4a4265bf4",
				"zh:2d81bfab3c6a9b897fa8fbb5256c9e5a944e6ecbf7f73a2a3e2b53a2c4fbcfc5",
				"zh:4cfc744cfc37aeeeecd82800c70e2591b38447af9e3c51bcbf06a5efe842ed65",
				"zh:734fbb81508b264f772a076338ddf1c7b25534d2007a1738a7d55587478ed258",
				"zh:9a5502c364f58073599fff8cdd8adc32e7f7bcd00a4d9b57d2fff678fd8a8319",
				"zh:9bc528f7e78dbfd106f94b741b68dedd3dd3d31c3defcddcc1972c8e52a6b7db",
				"zh:c30db03d877f9a7ae0c19d3fd338bbf95cdddbf6df1023709dbfa99689abac14",
				"zh:c51d4065145b8f4ca45fc9a0f3ca7f2d933bc0302af2eead74f3ce64a9221ae8",
				"zh:e23029fc7f81723795d7da770131adb1ce6f4d32f0a57eb75d47e036a0a19833",
				"zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
			},
			wantErr:    false,
			workingDir: "",
			expectedResult: `# This file is maintained automatically by "terraform init".
# Manual edits may be lost in future updates.

provider "registry.terraform.io/hashicorp/azuread" {
  version     = "2.41.0"
  constraints = "~> 2.39"
  hashes = [
    "h1:k3SUdcspQB5NtSux67Ngh7mkq/J3HW/6FllVhwK8vu8=",
    "zh:1c3e89cf19118fc07d7b04257251fc9897e722c16e0a0df7b07fcd261f8c12e7",
    "zh:32032e1539da9c986adea46a4fc702b8e3dea287ac825fdbc57ce459e6cfd25b",
    "zh:37841384780140f7026faf55700087d5e08f504168d67ae7f0d28e7132430c82",
    "zh:4c272ac7bf8b1aa913eeeca5fcdff25af432c6ae397d963710004b546a068936",
    "zh:8febf43db3f0bd9d46b2ee31eecf8f3ba912a4553ae049657e0f0c68bdff90a0",
    "zh:b60a3e92cd2a7af916d70d94d01c67d5bfabf68963f1b5ff75d8a310d66a4522",
    "zh:b9a8b9554f8ba9d306427aa8b2c0894242286972a61ab33300d6d8ac57bf93ab",
    "zh:ce7a79964a68a6086fa97e1f38d1c015d24d0732dd7876e179704c383a79f9aa",
    "zh:e305876a4d44739135264bb11be2c5489903682bb89f7b25561d38be31a9087d",
    "zh:ec3f63a848b3b2521cda0cf7650c4f85b50cc8905e5065f775d7c855833ab306",
    "zh:ef2a1ba97f15db88510bbc9b1af611c9bc8bee660d4d8b3b826424e1025487aa",
    "zh:f40db067d868567e199e16054a554aba5d89c3a19a4264c813dc7212724eeeea",
  ]
}

provider "registry.terraform.io/hashicorp/azurerm" {
  version     = "3.71.0"
  constraints = "~> 3.61"
  hashes = [
    "h1:QI0iaPNi0qAOIbXptd4ZObi0D5X1jojom5774GtEspA=",
    "zh:06f0d225b1711dfad256ff33134f878acc8f84624d9da66b075b075cc4d75892",
    "zh:09ff74056818babe02ea5a633bffe2b8223eaf79916dc1db169651ef7725c22f",
    "zh:27687e0f8458e6d88ebea94352eb523f56e8f5cdc468268af8f38dc4a4265bf4",
    "zh:2d81bfab3c6a9b897fa8fbb5256c9e5a944e6ecbf7f73a2a3e2b53a2c4fbcfc5",
    "zh:4cfc744cfc37aeeeecd82800c70e2591b38447af9e3c51bcbf06a5efe842ed65",
    "zh:734fbb81508b264f772a076338ddf1c7b25534d2007a1738a7d55587478ed258",
    "zh:9a5502c364f58073599fff8cdd8adc32e7f7bcd00a4d9b57d2fff678fd8a8319",
    "zh:9bc528f7e78dbfd106f94b741b68dedd3dd3d31c3defcddcc1972c8e52a6b7db",
    "zh:c30db03d877f9a7ae0c19d3fd338bbf95cdddbf6df1023709dbfa99689abac14",
    "zh:c51d4065145b8f4ca45fc9a0f3ca7f2d933bc0302af2eead74f3ce64a9221ae8",
    "zh:e23029fc7f81723795d7da770131adb1ce6f4d32f0a57eb75d47e036a0a19833",
    "zh:f569b65999264a9416862bca5cd2a6177d94ccb0424f3a4ef424428912b9cb3c",
  ]
}
`,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			err = h.Read()

			require.NoError(t, err)

			err = h.Apply(tt.spec.File, tt.value, tt.hashes)

			require.NoError(t, err)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, h.files[tt.spec.File].content)
		})
	}
}

func TestUpdateAbsoluteFilePath(t *testing.T) {
	testData := []struct {
		name           string
		spec           Spec
		workingDir     string
		expectedResult []string
	}{
		{
			name: "Success - Empty working directory",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			workingDir:     "",
			expectedResult: []string{"testdata/terraform.lock.hcl"},
		},
		{
			name: "Success - Working directory",
			spec: Spec{
				File:      "testdata/terraform.lock.hcl",
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			workingDir:     "/tmp",
			expectedResult: []string{"/tmp/testdata/terraform.lock.hcl"},
		},
		{
			name: "Success - Files with empty working directory",
			spec: Spec{
				Files:     []string{"testdata/terraform.lock.hcl", "testdata/data2.hcl"},
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			workingDir:     "",
			expectedResult: []string{"testdata/terraform.lock.hcl", "testdata/data2.hcl"},
		},
		{
			name: "Success - Working directory",
			spec: Spec{
				Files:     []string{"testdata/terraform.lock.hcl", "testdata/data2.hcl"},
				Provider:  "registry.terraform.io/hashicorp/kubernetes",
				Platforms: []string{"linux_amd64"},
			},
			workingDir:     "/tmp",
			expectedResult: []string{"/tmp/testdata/terraform.lock.hcl", "/tmp/testdata/data2.hcl"},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			h.UpdateAbsoluteFilePath(tt.workingDir)

			for _, v := range h.files {
				assert.True(t, slices.Contains(tt.expectedResult, v.filePath), fmt.Sprintf("%s not in %v", v.filePath, tt.expectedResult))
			}
		})
	}
}
