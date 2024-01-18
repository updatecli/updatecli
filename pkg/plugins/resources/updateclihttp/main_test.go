package updateclihttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		spec     Spec
		wantSpec Spec
		wantErr  bool
	}{
		{
			name: "Normal case with default index",
			spec: Spec{
				Url: "https://www.google.com",
			},
			wantSpec: Spec{
				Url: "https://www.google.com",
			},
		},
		{
			name: "Source to return response header instead of body with custom request",
			spec: Spec{
				Url:                  "https://www.google.com",
				ReturnResponseHeader: "Location",
				Request: Request{
					Verb: "POST",
					Body: multiLineText,
					Headers: map[string]string{
						"Authorization": "Bearer Token",
						"Accept":        "application/json",
					},
				},
			},
			wantSpec: Spec{
				Url:                  "https://www.google.com",
				ReturnResponseHeader: "Location",
				Request: Request{
					Verb: "POST",
					Body: multiLineText,
					Headers: map[string]string{
						"Authorization": "Bearer Token",
						"Accept":        "application/json",
					},
				},
			},
		},
		{
			name: "Condition with asserts on the response",
			spec: Spec{
				Url: "https://www.google.com",
				ResponseAsserts: ResponseAsserts{
					StatusCode: 404,
					Headers: map[string]string{
						"Content-Type":    "application/json",
						"x-frame-options": "DENY",
					},
				},
			},
			wantSpec: Spec{
				Url: "https://www.google.com",
				ResponseAsserts: ResponseAsserts{
					StatusCode: 404,
					Headers: map[string]string{
						"Content-Type":    "application/json",
						"x-frame-options": "DENY",
					},
				},
			},
		},
		{
			name: "Error when trying to create a resource with asserts (condition only) and a ReturnResponseHeader (source only)",
			spec: Spec{
				Url:                  "https://www.google.com",
				ReturnResponseHeader: "Content-Type",
				ResponseAsserts: ResponseAsserts{
					StatusCode: 404,
					Headers: map[string]string{
						"Content-Type":    "application/json",
						"x-frame-options": "DENY",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error when trying to create a resource without URL",
			spec: Spec{
				ReturnResponseHeader: "Content-Type",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := New(tt.spec)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantSpec, got.spec)
		})
	}
}
