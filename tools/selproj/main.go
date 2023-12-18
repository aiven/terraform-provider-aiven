//go:build tools

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client/v2"
)

// main is the entry point of the program.
func main() {
	token, ok := os.LookupEnv("AIVEN_TOKEN")
	if !ok {
		panic("AIVEN_TOKEN env var not set")
	}

	prefix, ok := os.LookupEnv("AIVEN_PROJECT_NAME_PREFIX")
	if !ok {
		panic("AIVEN_PROJECT_NAME_PREFIX env var not set")
	}

	client, err := aiven.NewTokenClient(token, "aiven/selproj")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	selectedProject, err := selectProject(ctx, NewAivenClientAdapter(client), prefix)
	if err != nil {
		panic(err)
	}

	fmt.Print(selectedProject)
}
