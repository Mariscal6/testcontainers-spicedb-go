package spicedb

import (
	"context"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

func (c *spiceDBContainer) GetEndpoint(_ context.Context) string {
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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("50051/tcp"),
		testcontainers.WithCmd("serve", "--grpc-preshared-key", defaultSecretKey),
		testcontainers.WithWaitStrategy(
			wait.ForAll(wait.ForExposedPort().WithPollInterval(2 * time.Second)),
		),
	}
	for _, opt := range opts {
		if secretKeyCustomizer, ok := opt.(SecretKeyCustomizer); ok {
			cfg.SecretKey = secretKeyCustomizer.SecretKey
		}

		if modelCustomizer, ok := opt.(ModelCustomizer); ok {
			cfg.Model = modelCustomizer.Model
		}
	}

	moduleOpts = append(moduleOpts, opts...)
	container, err := testcontainers.Run(ctx, image, moduleOpts...)
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
		req.Cmd = append(req.Cmd, "--http-enabled", "--http-addr", ":"+port)
		req.ExposedPorts = append(req.ExposedPorts, port+"/tcp")
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
		PostReadies: []testcontainers.ContainerHook{
			func(ctx context.Context, c testcontainers.Container) error {
				if customizer.SchremaWriter == nil {
					return customizer.defaultSchemaWriterfunc(ctx, c)
				}
				return customizer.SchremaWriter(ctx, c, customizer.Model, customizer.SecretKey)
			},
		},
	})
	return nil
}

func (customizer ModelCustomizer) defaultSchemaWriterfunc(ctx context.Context, c testcontainers.Container) error {
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

	_, err = client.WriteSchema(ctx, &v1.WriteSchemaRequest{
		Schema: customizer.Model,
	})
	return err
}
