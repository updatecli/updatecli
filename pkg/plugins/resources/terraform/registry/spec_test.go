package registry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Failure - Type missing",
			spec: Spec{
				Type:      "",
				Namespace: "hashicorp",
				Name:      "kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type wrong",
			spec: Spec{
				Type:      "type",
				Namespace: "hashicorp",
				Name:      "kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Success - Type provider",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				Name:      "kubernetes",
			},
			wantErr: false,
		},
		{
			name: "Success - Type provider raw string",
			spec: Spec{
				Type:      "provider",
				RawString: "hashicorp/kubernetes",
			},
			wantErr: false,
		},
		{
			name: "Success - Type provider raw string with hostname",
			spec: Spec{
				Type:      "provider",
				RawString: "registry.terraform.io/hashicorp/kubernetes",
			},
			wantErr: false,
		},
		{
			name: "Failure - Type provider without namespace",
			spec: Spec{
				Type: "provider",
				Name: "kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type provider without name",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Success - Type provider",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				Name:      "kubernetes",
			},
			wantErr: false,
		},
		{
			name: "Failure - Type provider raw string and all other fields",
			spec: Spec{
				Type:      "provider",
				Hostname:  "registry.terraform.io",
				Namespace: "hashicorp",
				Name:      "kubernetes",
				RawString: "registry.terraform.io/hashicorp/kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module raw string and hostname",
			spec: Spec{
				Type:      "provider",
				Hostname:  "registry.terraform.io",
				RawString: "registry.terraform.io/hashicorp/kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type provider raw string and namespace",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				RawString: "registry.terraform.io/hashicorp/kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type provider raw string and name",
			spec: Spec{
				Type:      "provider",
				Name:      "kubernetes",
				RawString: "registry.terraform.io/hashicorp/kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type provider and target system",
			spec: Spec{
				Type:         "provider",
				Namespace:    "hashicorp",
				Name:         "kubernetes",
				TargetSystem: "aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Success - Type module",
			spec: Spec{
				Type:         "module",
				Namespace:    "terraform-aws-modules",
				Name:         "vpc",
				TargetSystem: "aws",
			},
			wantErr: false,
		},
		{
			name: "Success - Type provider raw string",
			spec: Spec{
				Type:      "provider",
				RawString: "terraform-aws-modules/vpc/aws",
			},
			wantErr: false,
		},
		{
			name: "Success - Type provider raw string with hostname",
			spec: Spec{
				Type:      "provider",
				RawString: "app.terraform.io/terraform-aws-modules/vpc/aws",
			},
			wantErr: false,
		},
		{
			name: "Failure - Type module with namespace",
			spec: Spec{
				Type:         "module",
				Name:         "vpc",
				TargetSystem: "aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module with name",
			spec: Spec{
				Type:         "module",
				Namespace:    "terraform-aws-modules",
				TargetSystem: "aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module with target system",
			spec: Spec{
				Type:      "module",
				Namespace: "terraform-aws-modules",
				Name:      "vpc",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module raw string and all other fields",
			spec: Spec{
				Type:         "module",
				Hostname:     "app.terraform.io",
				Namespace:    "terraform-aws-modules",
				Name:         "vpc",
				TargetSystem: "aws",
				RawString:    "app.terraform.io/terraform-aws-modules/vpc/aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module raw string and hostname",
			spec: Spec{
				Type:      "module",
				Hostname:  "app.terraform.io",
				RawString: "app.terraform.io/terraform-aws-modules/vpc/aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module raw string and namespace",
			spec: Spec{
				Type:      "module",
				Namespace: "terraform-aws-modules",
				RawString: "app.terraform.io/terraform-aws-modules/vpc/aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module raw string and name",
			spec: Spec{
				Type:      "module",
				Name:      "vpc",
				RawString: "app.terraform.io/terraform-aws-modules/vpc/aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Type module raw string and target system",
			spec: Spec{
				Type:         "module",
				TargetSystem: "aws",
				RawString:    "app.terraform.io/terraform-aws-modules/vpc/aws",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
