package lambdamiddleware_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mashiike/lambdamiddleware"
)

func Example() {
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
		func(next lambda.Handler) lambda.Handler {
			return lambdamiddleware.HandlerFunc(
				func(ctx context.Context, payload []byte) ([]byte, error) {
					fmt.Println("[start middleware2]")
					output, err := next.Invoke(ctx, payload)
					fmt.Println("[end middleware2]")
					return output, err
				},
			)
		},
	)
	s.Append(
		func(next lambda.Handler) lambda.Handler {
			return lambdamiddleware.HandlerFunc(
				func(ctx context.Context, payload []byte) ([]byte, error) {
					fmt.Println("[start middleware3]")
					output, err := next.Invoke(ctx, payload)
					fmt.Println("[end middleware3]")
					return output, err
				},
			)
		},
	)
	handler := s.Then(func(ctx context.Context, event json.RawMessage) error {
		fmt.Println("[handler]", string(event))
		return nil
	})
	// lambda.Start(handler)
	handler.Invoke(context.Background(), []byte("{}"))

	//output:
	//[start middleware1]
	//[start middleware2]
	//[start middleware3]
	//[handler] {}
	//[end middleware3]
	//[end middleware2]
	//[end middleware1]
}
