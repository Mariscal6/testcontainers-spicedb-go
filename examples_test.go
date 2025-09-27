package spicedb_test

import (
	"context"
	"fmt"
	"log"

	spicedb "github.com/Mariscal6/testcontainers-spicedb-go"
)

func ExampleRunContainer() {
	// runspiceDBContainer {
	ctx := context.Background()

	spicedbContainer, err := spicedb.Run(ctx, "authzed/spicedb:v1.33.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := spicedbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := spicedbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) //nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
