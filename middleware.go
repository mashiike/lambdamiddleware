package lambdamiddleware

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

// HandlerFunc er en type anonym funktion, der opfylder grænsefladen lambda.Handler
type HandlerFunc func(ctx context.Context, payload []byte) ([]byte, error)

func (h HandlerFunc) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return h(ctx, payload)
}

// Middleware describes the processing that precedes the execution of the Lambda Handler.
type Middleware func(next lambda.Handler) lambda.Handler

// Stack is a middleware Stack for the Lambda Handler.
type Stack struct {
	middlewares []Middleware
}

func NewStack(middlewares ...Middleware) *Stack {
	return &Stack{
		middlewares: append([]Middleware(nil), middlewares...),
	}
}

// Append adds new middleware to Stack The added middleware is executed last.
func (s *Stack) Append(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

// Then apply the middleware to a certain handler.
func (s *Stack) Then(handler interface{}) lambda.Handler {
	h, ok := handler.(lambda.Handler)
	if !ok {
		h = lambda.NewHandler(handler)
	}
	for i := range s.middlewares {
		h = s.middlewares[len(s.middlewares)-1-i](h)
	}
	return h
}
