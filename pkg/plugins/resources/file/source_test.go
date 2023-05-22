package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_Source(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		files          map[string]fileMetadata
		mockedContents map[string]string
		mockedError    error
		wantedContents map[string]string
		wantedResult   bool
		wantedErr      bool
	}{
		{
			name: "Passing case with 'File'",
			spec: Spec{
				File: "/home/ucli/foo.txt",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/home/ucli/foo.txt": "Hello World",
			},
			wantedContents: map[string]string{
				"/home/ucli/foo.txt": "Hello World",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Files'",
			spec: Spec{
				Files: []string{
					"/home/ucli/foo.txt",
				},
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/home/ucli/foo.txt": "Hello World",
			},
			wantedContents: map[string]string{
				"/home/ucli/foo.txt": "Hello World",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'File' and 'Line' specified",
			spec: Spec{
				File: "/home/ucli/foo.txt",
				Line: 3,
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/home/ucli/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedContents: map[string]string{
				"/home/ucli/foo.txt": "The End",
			},
		},
		{
			name: "Passing case with single-line 'MatchPattern'",
			spec: Spec{
				File:         "/home/ucli/foo.txt",
				MatchPattern: ".*freebsd_386.*",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/home/ucli/foo.txt": `363d0e0c5c4cb4e69f5f2c7f64f9bf01ab73af0801665d577441521a24313a07  terraform_0.14.5_darwin_amd64.zip
					5a3e0c7873faa048f59d563a2a98caf7f04045967cbb5ad6cf05f5991e20b8d1  terraform_0.14.5_freebsd_386.zip
					4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip
					03c201a9a3e1d2776d0cfc0163e52484f3dbbbd73eb08d5bac491ca87a9aa3b7  terraform_0.14.5_freebsd_arm.zip
					b262998c85a7cad1c24b90f3d309d592bd349d411167a2939eb482dc2b99702d  terraform_0.14.5_linux_386.zip
					2899f47860b7752e31872e4d57b1c03c99de154f12f0fc84965e231bc50f312f  terraform_0.14.5_linux_amd64.zip
					a971a5f5da82ea896a2e91fd828c90ea9c28e3de575d03a7ce25a5840ed7ae2b  terraform_0.14.5_linux_arm.zip
					d3cab7d777eec230b67eb9723f3b271cd43e29c688439e4c67e3398cdaf6406b  terraform_0.14.5_linux_arm64.zip
					67b153c8c754ca03e3f8954b201cf27ec31387c8d3c77d302d647417bc4a23f4  terraform_0.14.5_openbsd_386.zip
					062fbc3f596490e33e6493a8e186ae50e7b6077ac2a842392991d918189187fc  terraform_0.14.5_openbsd_amd64.zip
					f66920ffedd7e81cd116d185ada479ba466f5514f8b20194cc180d3c6184e060  terraform_0.14.5_solaris_amd64.zip
					f8bf1fca0ef11a33955d225198d1211e15827d43488cc9174dcda14d1a7a1d19  terraform_0.14.5_windows_386.zip
					5d25f9afc71fc49d5f3e8c7ccc3ccd83a840c56e7a015f55f321fc970a73050b  terraform_0.14.5_windows_amd64.zip`,
			},
			wantedContents: map[string]string{
				"/home/ucli/foo.txt": "					5a3e0c7873faa048f59d563a2a98caf7f04045967cbb5ad6cf05f5991e20b8d1  terraform_0.14.5_freebsd_386.zip",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with multi-line 'MatchPattern'",
			spec: Spec{
				File:         "/home/ucli/foo.txt",
				MatchPattern: ".*terraform_.*_linux_.*",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/home/ucli/foo.txt": `363d0e0c5c4cb4e69f5f2c7f64f9bf01ab73af0801665d577441521a24313a07  terraform_0.14.5_darwin_amd64.zip
					5a3e0c7873faa048f59d563a2a98caf7f04045967cbb5ad6cf05f5991e20b8d1  terraform_0.14.5_freebsd_386.zip
					4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip
					03c201a9a3e1d2776d0cfc0163e52484f3dbbbd73eb08d5bac491ca87a9aa3b7  terraform_0.14.5_freebsd_arm.zip
					b262998c85a7cad1c24b90f3d309d592bd349d411167a2939eb482dc2b99702d  terraform_0.14.5_linux_386.zip
					2899f47860b7752e31872e4d57b1c03c99de154f12f0fc84965e231bc50f312f  terraform_0.14.5_linux_amd64.zip
					a971a5f5da82ea896a2e91fd828c90ea9c28e3de575d03a7ce25a5840ed7ae2b  terraform_0.14.5_linux_arm.zip
					d3cab7d777eec230b67eb9723f3b271cd43e29c688439e4c67e3398cdaf6406b  terraform_0.14.5_linux_arm64.zip
					67b153c8c754ca03e3f8954b201cf27ec31387c8d3c77d302d647417bc4a23f4  terraform_0.14.5_openbsd_386.zip
					062fbc3f596490e33e6493a8e186ae50e7b6077ac2a842392991d918189187fc  terraform_0.14.5_openbsd_amd64.zip
					f66920ffedd7e81cd116d185ada479ba466f5514f8b20194cc180d3c6184e060  terraform_0.14.5_solaris_amd64.zip
					f8bf1fca0ef11a33955d225198d1211e15827d43488cc9174dcda14d1a7a1d19  terraform_0.14.5_windows_386.zip
					5d25f9afc71fc49d5f3e8c7ccc3ccd83a840c56e7a015f55f321fc970a73050b  terraform_0.14.5_windows_amd64.zip`,
			},
			wantedContents: map[string]string{
				"/home/ucli/foo.txt": `					b262998c85a7cad1c24b90f3d309d592bd349d411167a2939eb482dc2b99702d  terraform_0.14.5_linux_386.zip
					2899f47860b7752e31872e4d57b1c03c99de154f12f0fc84965e231bc50f312f  terraform_0.14.5_linux_amd64.zip
					a971a5f5da82ea896a2e91fd828c90ea9c28e3de575d03a7ce25a5840ed7ae2b  terraform_0.14.5_linux_arm.zip
					d3cab7d777eec230b67eb9723f3b271cd43e29c688439e4c67e3398cdaf6406b  terraform_0.14.5_linux_arm64.zip`,
			},
		},
		{
			name: "Validation failure with more than one element in 'Files'",
			spec: Spec{
				Files: []string{
					"/home/ucli/foo.txt",
					"/home/ucli/bar.txt",
				},
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
				"/home/ucli/bar.txt": {
					originalPath: "/home/ucli/bar.txt",
					path:         "/home/ucli/bar.txt",
				},
			},
			wantedErr: true,
		},
		{
			name: "Validation failure with 'ReplacePattern' specified",
			spec: Spec{
				MatchPattern:   "maven_(.*)",
				ReplacePattern: "gradle_$1",
				File:           "/home/ucli/foo.txt",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			wantedErr: true,
		},
		{
			name: "Validation failure with 'Content' specified",
			spec: Spec{
				Content: "Hello world",
				File:    "/home/ucli/foo.txt",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			wantedErr: true,
		},
		{
			name: "Validation failure with 'ForceCreate' specified",
			spec: Spec{
				ForceCreate: true,
				File:        "/home/ucli/foo.txt",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			wantedErr: true,
		},
		{
			name: "Validation failure with invalid regexp 'MatchPattern' specified",
			spec: Spec{
				MatchPattern: "(d+:1",
				File:         "/home/ucli/foo.txt",
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			wantedErr: true,
		},
		{
			name:        "Failing case with nonexistent 'File'",
			files:       map[string]fileMetadata{},
			mockedError: fmt.Errorf("no such file or directory"),
			wantedErr:   true,
		},
		{
			name: "Failing case with 'File' and nonexistent 'Line' specified",
			spec: Spec{
				File: "/home/ucli/foo.txt",
				Line: 3,
			},
			files: map[string]fileMetadata{
				"/home/ucli/foo.txt": {
					originalPath: "/home/ucli/foo.txt",
					path:         "/home/ucli/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/home/ucli/foo.txt": "Don't worry\r\nBe happy",
			},
			wantedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedText := text.MockTextRetriever{
				Contents: tt.mockedContents,
				Err:      tt.mockedError,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockedText,
				files:            tt.files,
			}
			// Looping on the only filePath in 'files'
			for filePath := range f.files {
				gotResult := result.Source{}
				gotErr := f.Source(filePath, &gotResult)
				if tt.wantedErr {
					assert.Error(t, gotErr)
					return
				}

				require.NoError(t, gotErr)
				assert.Equal(t, tt.wantedContents[filePath], gotResult.Information)
			}
		})
	}
}
