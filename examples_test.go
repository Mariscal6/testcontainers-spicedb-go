package spicedb_test

import (
	"context"
	"fmt"
	"log"

	spicedb "github.com/mariscal6/testcontainers-spicedb-go"

	"github.com/testcontainers/testcontainers-go"
)

func ExampleRunContainer() {
	// runspiceDBContainer {
	ctx := context.Background()

	spicedbContainer, err := spicedb.RunContainer(ctx, testcontainers.WithImage("authzed/spicedb:v1.33.0"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := spicedbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := spicedbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
