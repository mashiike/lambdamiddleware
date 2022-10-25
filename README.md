# lambdamiddleware
Useful lambda handler middleware repository


```go
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mashiike/lambdamiddleware"
)

func main() {
    s := lambdamiddleware.NewStack(
        func(next lambda.Handler) lambda.Handler {
            return lambdamiddleware.HandlerFunc(
                func(ctx context.Context, payload []byte) ([]byte, error) {
                    fmt.Println("[start middleware1]")
                    output, err := next.Invoke(ctx, payload)
                    fmt.Println("[end middleware1]")
                    return output, err
                },
            )
        },
    )
    handler := s.Then(func(ctx context.Context, event json.RawMessage) error {
        fmt.Println("[handler]", string(event))
        return nil
    })
    lambda.Start(handler)
}
```

### Parameters And Secrets Extention Hepler Middleware

helper middleware for https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html

```go
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
```

see deteils [exampels/parameters-and-secrets](_examples/parameters-and-secrets)

## LICENSE 

MIT
