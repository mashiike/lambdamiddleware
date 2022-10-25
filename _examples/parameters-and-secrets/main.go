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

func main() {
	paramsAndSecrets, err := lambdamiddleware.ParametersAndSecrets(&lambdamiddleware.ParametersAndSecretsConfig{
		Names: strings.Split(os.Getenv("SSMNAMES"), ","),
	})
	if err != nil {
		log.Fatalln(err)
	}
	lambda.Start(lambdamiddleware.NewStack(paramsAndSecrets).Then(
		func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
			return map[string]interface{}{
				"foo":  ctx.Value("/lambdamiddleware-examples/foo"),
				"bar":  ctx.Value("/lambdamiddleware-examples/bar"),
				"tora": ctx.Value("/lambdamiddleware-examples/tora"),
			}, nil
		},
	))
}
