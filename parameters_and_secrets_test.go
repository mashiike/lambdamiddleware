package lambdamiddleware_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/mashiike/lambdamiddleware"
	"github.com/stretchr/testify/require"
)

func TestParametersAndSecrets(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/systemsmanager/parameters/get" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Aws-Parameters-Secrets-Token") != "SessionToken" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		name := r.URL.Query().Get("name")
		output := &ssm.GetParameterOutput{
			Parameter: &types.Parameter{
				ARN:              aws.String(fmt.Sprintf("arn:aws:ssm:ap-northeast-1:0123456789012:parameter/%s", strings.TrimPrefix(name, "/"))),
				Name:             aws.String(name),
				DataType:         aws.String("text"),
				LastModifiedDate: aws.Time(time.Now()),
				Type:             types.ParameterTypeString,
				Value:            aws.String("dummy_value"),
				Version:          1,
			},
		}
		json.NewEncoder(w).Encode(output)
	}))
	defer testServer.Close()
	t.Log(testServer.URL)
	u, err := url.Parse(testServer.URL)
	require.NoError(t, err)
	port, err := strconv.ParseInt(u.Port(), 10, 64)
	require.NoError(t, err)

	cases := []struct {
		casename            string
		sessionToken        string
		handler             interface{}
		cfg                 *lambdamiddleware.ParametersAndSecretsConfig
		payload             []byte
		output              string
		initializeErrString string
		errString           string
	}{
		{
			casename: "AWS_SESSION_TOKEN not set",
			handler: func(_ context.Context, payload json.RawMessage) error {
				require.EqualValues(t, `"hoge"`, string(payload))
				return nil
			},
			payload:             []byte(`"hoge"`),
			initializeErrString: "AWS_SESSION_TOKEN not set",
		},
		{
			casename:     "success",
			sessionToken: "SessionToken",
			cfg: &lambdamiddleware.ParametersAndSecretsConfig{
				ExtentionHTTPPort: int(port),
				Names:             []string{"hoge", "fuga"},
				ContextKeyFunc: func(key string) interface{} {
					return "ssm:" + key
				},
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				require.EqualValues(t, `"hoge"`, string(payload))
				return map[string]interface{}{
					"hoge": ctx.Value("ssm:hoge"),
				}, nil
			},
			payload: []byte(`"hoge"`),
			output:  `{"hoge":"dummy_value"}`,
		},
		{
			casename:     "invalid session token",
			sessionToken: "InvalidSessionToken",
			cfg: &lambdamiddleware.ParametersAndSecretsConfig{
				ExtentionHTTPPort: int(port),
				Names:             []string{"hoge", "fuga"},
				ContextKeyFunc: func(key string) interface{} {
					return "ssm:" + key
				},
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				require.EqualValues(t, `"hoge"`, string(payload))
				return map[string]interface{}{
					"hoge": ctx.Value("ssm:hoge"),
				}, nil
			},
			payload:   []byte(`"hoge"`),
			errString: "HTTP Status 401: 401 Unauthorized",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case.%d:%s", i+1, c.casename), func(t *testing.T) {
			os.Setenv("AWS_SESSION_TOKEN", c.sessionToken)
			h, err := lambdamiddleware.WrapParametersAndSecrets(c.handler, c.cfg)
			if c.initializeErrString == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, c.initializeErrString)
				return
			}
			output, err := h.Invoke(context.Background(), c.payload)
			if c.errString == "" {
				require.NoError(t, err)
				require.JSONEq(t, c.output, string(output))
			} else {
				require.EqualError(t, err, c.errString)
			}
		})
	}
}
