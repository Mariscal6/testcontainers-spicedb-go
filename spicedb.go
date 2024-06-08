package spicedb

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// spiceDBContainer represents the spiceDB container type used in the module
type spiceDBContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the spiceDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*spiceDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "authzed/spicedb:v1.33.0",
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &spiceDBContainer{Container: container}, nil
}
