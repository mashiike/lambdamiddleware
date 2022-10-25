package lambdamiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"golang.org/x/sync/errgroup"
)

type ParametersAndSecretsConfig struct {
	Names             []string
	ExtentionHTTPPort int
	Client            *http.Client
	ContextKeyFunc    func(key string) interface{}
	SetEnv            bool
	EnvPrefix         string
	withDecryption    *bool
}

func (cfg *ParametersAndSecretsConfig) WithDecryption(value bool) *ParametersAndSecretsConfig {
	cfg.withDecryption = &value
	return cfg
}

func WrapParametersAndSecrets(handler interface{}, cfg *ParametersAndSecretsConfig) (lambda.Handler, error) {
	m, err := ParametersAndSecrets(cfg)
	if err != nil {
		return nil, err
	}
	s := NewStack(m)
	return s.Then(handler), nil
}

func ParametersAndSecrets(cfg *ParametersAndSecretsConfig) (Middleware, error) {
	if cfg == nil {
		cfg = &ParametersAndSecretsConfig{}
	}
	if cfg.ExtentionHTTPPort == 0 {
		if extentionHTTPPortStr := os.Getenv("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT"); extentionHTTPPortStr != "" {
			extentionHTTPPort, err := strconv.ParseInt(extentionHTTPPortStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("can not PARAMETERS_SECRETS_EXTENSION_HTTP_PORT parse as int:%w", err)
			}
			cfg.ExtentionHTTPPort = int(extentionHTTPPort)
		} else {
			cfg.ExtentionHTTPPort = 2773
		}
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	if cfg.withDecryption == nil {
		cfg = cfg.WithDecryption(true)
	}
	if cfg.ContextKeyFunc == nil {
		cfg.ContextKeyFunc = func(key string) interface{} {
			return key
		}
	}
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	if sessionToken == "" {
		return nil, errors.New("AWS_SESSION_TOKEN not set")
	}
	if cfg.SetEnv {
		_, err := fetchParametersAndSecrets(context.Background(), sessionToken, cfg)
		if err != nil {
			return nil, err
		}
	}
	return func(next lambda.Handler) lambda.Handler {
		return HandlerFunc(func(ctx context.Context, payload []byte) ([]byte, error) {
			ctx, err := fetchParametersAndSecrets(ctx, sessionToken, cfg)
			if err != nil {
				return nil, err
			}
			return next.Invoke(ctx, payload)
		})
	}, nil
}

func fetchParametersAndSecrets(ctx context.Context, sessionToken string, cfg *ParametersAndSecretsConfig) (context.Context, error) {
	eg, egctx := errgroup.WithContext(ctx)
	var m sync.Map
	for _, name := range cfg.Names {
		_name := name
		eg.Go(func() error {
			u := &url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("localhost:%d", cfg.ExtentionHTTPPort),
				Path:   "/systemsmanager/parameters/get",
				RawQuery: url.Values{
					"name":           []string{_name},
					"withDecryption": []string{fmt.Sprintf("%v", *cfg.withDecryption)},
				}.Encode(),
			}
			req, err := http.NewRequestWithContext(egctx, http.MethodGet, u.String(), nil)
			if err != nil {
				return err
			}
			req.Header.Set("X-Aws-Parameters-Secrets-Token", sessionToken)
			resp, err := cfg.Client.Do(req)
			if err != nil {
				return err
			}
			defer func() {
				io.ReadAll(resp.Body)
				resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("HTTP Status %d: %s", resp.StatusCode, resp.Status)
			}
			decoder := json.NewDecoder(resp.Body)
			var output ssm.GetParameterOutput
			if err := decoder.Decode(&output); err != nil {
				return err
			}
			m.Store(_name, *output.Parameter.Value)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return ctx, err
	}
	m.Range(func(key, value interface{}) bool {
		ctx = context.WithValue(ctx, cfg.ContextKeyFunc(key.(string)), value.(string))
		if cfg.SetEnv {
			parts := strings.Split(key.(string), "/")
			envKey := strings.ToUpper(cfg.EnvPrefix + parts[len(parts)-1])
			os.Setenv(envKey, value.(string))
		}
		return true
	})
	return ctx, nil
}
