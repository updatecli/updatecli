package temurin

// import (
// 	"net/http"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
// )

// func TestCondition(t *testing.T) {
// 	tests := []struct {
// 		name                  string
// 		spec                  Spec
// 		source                string
// 		scm                   scm.ScmHandler
// 		mockedHTTPStatusCode  int
// 		mockedHTTPBody        string
// 		mockedHTTPRespHeaders http.Header
// 		mockedHttpError       error
// 		want                  bool
// 		wantErr               error
// 	}{
// 		{
// 			name:                 "Success case with existing URL",
// 			spec:                 Spec{},
// 			mockedHTTPStatusCode: http.StatusOK,
// 			want:                 true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			sut, sutErr := New(tt.spec)
// 			require.NoError(t, sutErr)

// 			// sut.httpClient = &httpclient.MockClient{
// 			// 	DoFunc: func(req *http.Request) (*http.Response, error) {
// 			// 		body := tt.mockedHTTPBody
// 			// 		statusCode := tt.mockedHTTPStatusCode
// 			// 		return &http.Response{
// 			// 			StatusCode: statusCode,
// 			// 			Body:       io.NopCloser(strings.NewReader(body)),
// 			// 			Header:     tt.mockedHTTPRespHeaders,
// 			// 		}, tt.mockedHttpError
// 			// 	},
// 			// }

// 			got, _, gotErr := sut.Condition(tt.source, tt.scm)

// 			if tt.wantErr != nil {
// 				require.Error(t, gotErr)
// 				assert.Equal(t, tt.wantErr, gotErr)
// 				return
// 			}

// 			require.NoError(t, gotErr)
// 			assert.Equal(t, tt.want, got)
// 		})
// 	}
// }
