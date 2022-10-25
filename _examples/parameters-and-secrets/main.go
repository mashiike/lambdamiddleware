package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mashiike/lambdamiddleware"
)

type ssmContextKey string

func main() {
	paramsAndSecrets, err := lambdamiddleware.ParametersAndSecrets(&lambdamiddleware.ParametersAndSecretsConfig{
		Names:          strings.Split(os.Getenv("SSMNAMES"), ","),
		ContextKeyFunc: func(key string) interface{} { return ssmContextKey(key) },
		EnvPrefix:      "SSM_",
		SetEnv:         true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	lambda.Start(lambdamiddleware.NewStack(paramsAndSecrets).Then(
		func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
			return map[string]interface{}{
				"env_foo": os.Getenv("SSM_FOO"),
				"env_bar": os.Getenv("SSM_BAR"),
				"foo":     ctx.Value(ssmContextKey("/lambdamiddleware-examples/foo")),
				"bar":     ctx.Value(ssmContextKey("/lambdamiddleware-examples/bar")),
				"tora":    ctx.Value(ssmContextKey("/lambdamiddleware-examples/tora")),
			}, nil
		},
	))
}
