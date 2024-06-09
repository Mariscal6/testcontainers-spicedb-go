package spicedb

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultSecretKey = "somepresharedkey"
)

// spiceDBContainer represents the spiceDB container type used in the module
type Config struct {
	SecretKey string
	Model     string
}

type spiceDBContainer struct {
	testcontainers.Container
	secretKey string
	model     string
	endpoint  string
}

func (c *spiceDBContainer) SecretKey() string {
	return c.secretKey
}

func (c *spiceDBContainer) GetEndpoint(ctx context.Context) string {
	return c.endpoint
}

// RunContainer creates an instance of the spiceDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*spiceDBContainer, error) {
	cfg := Config{
		SecretKey: defaultSecretKey,
	}
	req := testcontainers.ContainerRequest{
		Image:        "authzed/spicedb:v1.33.0",
		ExposedPorts: []string{"50051/tcp"},
		Cmd:          []string{"serve", "--grpc-preshared-key", defaultSecretKey},
		WaitingFor: wait.ForAll(
			wait.ForLog("http server started serving"),
			// TODO: add a health grpc healthz check
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}

		if secretKeyCustomizer, ok := opt.(SecretKeyCustomizer); ok {
			cfg.SecretKey = secretKeyCustomizer.SecretKey
		}

		if modelCustomizer, ok := opt.(ModelCustomizer); ok {
			cfg.Model = modelCustomizer.Model
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		container.Terminate(ctx)
		return nil, err
	}
	return &spiceDBContainer{Container: container, secretKey: cfg.SecretKey, endpoint: endpoint, model: cfg.Model}, nil
}

func WithOtel(otelProvider string, enpoint string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, "--otel-endpoint", enpoint, "--otel-provider", otelProvider)
		return nil
	}
}

func WithHTTP(port string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, "--http-enabled", "--http-addr", fmt.Sprintf(":%s", port))
		req.ExposedPorts = append(req.ExposedPorts, fmt.Sprintf("%s/tcp", port))
		return nil
	}
}

type SecretKeyCustomizer struct {
	SecretKey string
}

func (customizer SecretKeyCustomizer) Customize(req *testcontainers.GenericContainerRequest) error {
	for i, cmd := range req.Cmd {
		if cmd == "--grpc-preshared-key" {
			req.Cmd[i+1] = customizer.SecretKey
			return nil
		}
	}
	req.Cmd = append(req.Cmd, "--grpc-preshared-key", customizer.SecretKey)
	return nil
}

type ModelCustomizer struct {
	Model     string
	SecretKey string
}

// Customize method implementation
func (customizer ModelCustomizer) Customize(req *testcontainers.GenericContainerRequest) error {
	req.LifecycleHooks = append(req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{
		PostStarts: []testcontainers.ContainerHook{
			func(ctx context.Context, c testcontainers.Container) error {
				// replace with a health check
				time.Sleep(2 * time.Second)
				endpoint, err := c.Endpoint(ctx, "")
				if err != nil {
					return err
				}

				client, err := authzed.NewClient(
					endpoint,
					grpcutil.WithInsecureBearerToken(customizer.SecretKey),
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
				if err != nil {
					return err
				}
				_, err = client.SchemaServiceClient.WriteSchema(ctx, &v1.WriteSchemaRequest{
					Schema: customizer.Model,
				})
				return err
			},
		},
	})
	return nil
}
