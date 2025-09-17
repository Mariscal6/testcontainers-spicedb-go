package spicedb

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

// Deprecated: use Run instead
// RunContainer creates an instance of the spiceDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*spiceDBContainer, error) {
	return Run(ctx, "authzed/spicedb:v1.33.0", opts...)
}

// Run creates an instance of the spiceDB container type
func Run(ctx context.Context, image string, opts ...testcontainers.ContainerCustomizer) (*spiceDBContainer, error) {
	cfg := Config{
		SecretKey: defaultSecretKey,
	}
	req := testcontainers.ContainerRequest{
		Image:        image,
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

	c := &spiceDBContainer{Container: container, secretKey: cfg.SecretKey, model: cfg.Model}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return c, err
	}

	c.endpoint = endpoint

	return c, nil
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
	Model         string
	SecretKey     string
	SchremaWriter func(ctx context.Context, c testcontainers.Container, model string, secret string) error
}

// Customize method implementation
func (customizer ModelCustomizer) Customize(req *testcontainers.GenericContainerRequest) error {
	req.LifecycleHooks = append(req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{
		PostStarts: []testcontainers.ContainerHook{
			func(ctx context.Context, c testcontainers.Container) error {
				return customizer.SchremaWriter(ctx, c, customizer.Model, customizer.SecretKey)
			},
		},
	})
	return nil
}
