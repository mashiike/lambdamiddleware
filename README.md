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
````
