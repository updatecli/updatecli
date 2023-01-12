package cargopackage

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	dir, err := CreateDummyIndex()
	defer os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}

	tests := []struct {
		name           string
		url            string
		spec           Spec
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "Retrieving existing rand version from the default index api",
			spec: Spec{
				Package: "rand",
				Version: "0.7.2",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "Retrieving non-existing rand version from the default index api",
			spec: Spec{
				Package: "rand",
				Version: "99.99.99",
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "Retrieving existing crate-test version from the filesystem index",
			spec: Spec{
				IndexDir: dir,
				Package:  "crate-test",
				Version:  "0.2.2",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "Retrieving existing yanked crate-test version from the filesystem index",
			spec: Spec{
				IndexDir: dir,
				Package:  "crate-test",
				Version:  "0.2.3",
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "Retrieving non-existing yanked crate-test version from the filesystem index",
			spec: Spec{
				IndexDir: dir,
				Package:  "crate-test",
				Version:  "99.99.99",
			},
			expectedResult: false,
			expectedError:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec, "")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			gotVersion, err := got.Condition("")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}
}
