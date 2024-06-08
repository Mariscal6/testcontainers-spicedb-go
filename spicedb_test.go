package spicedb_test

import (
	"context"
	"testing"

	spicedb "github.com/mariscal6/testcontainers-spicedb-go"
	"github.com/testcontainers/testcontainers-go"
)

func TestSpiceDB(t *testing.T) {
	ctx := context.Background()

	container, err := spicedb.RunContainer(ctx, testcontainers.WithImage("authzed/spicedb:v1.33.0"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
}
