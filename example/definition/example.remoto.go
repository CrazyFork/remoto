package example

import "context"

// Greeter provides greeting services.
type Greeter interface {
	// Greet generates a greeting.
	Greet(context.Context, *GreetRequest) (*GreetResponse, error)
}

// GreetFormatter provides formattable greeting services.
type GreetFormatter interface {
	// Greet generates a greeting.
	Greet(context.Context, *GreetFormatRequest) (*GreetResponse, error)
}

// GreetRequest is the request for Greeter.Greet.
type GreetRequest struct {
	Name string
}

// GreetResponse is the response for Greeter.Greet and GreetFormatter.Greet.
type GreetResponse struct {
	Greeting string
}

// GreetFormatRequest is the request for GreetFormatter.Greet.
type GreetFormatRequest struct {
	Format string
	Name   string
}
