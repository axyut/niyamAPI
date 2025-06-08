package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// GreetingOutput represents the greeting operation response.
type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func main() {
	// Create a new router & API.
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

	// Register GET /greeting/{name} handler.
	huma.Get(api, "/greeting/{name}", func(ctx context.Context, input *struct {
		Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
	}) (*GreetingOutput, error) {
		resp := &GreetingOutput{}
		resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)
		return resp, nil
	})

	port := "7860" // Default port
	// You might read the port from an environment variable for flexibility
	// if p := os.Getenv("PORT"); p != "" {
	//     port = p
	// }

	addr := fmt.Sprintf(":%s", port) // Listen on all interfaces
	fmt.Printf("Server starting on http://localhost%s/docs\n", addr)
	err := http.ListenAndServe(addr, router) // <--- THIS IS KEY
	if err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
