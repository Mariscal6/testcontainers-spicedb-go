package spicedb_test

import (
	"context"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	spicedbcontainer "github.com/mariscal6/testcontainers-spicedb-go"
	"github.com/mariscal6/testcontainers-spicedb-go/testdata"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSpiceDB(t *testing.T) {
	ctx := context.Background()

	container, err := spicedbcontainer.RunContainer(ctx,
		testcontainers.WithImage("authzed/spicedb:v1.33.0"),
	)

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	spicedbClient, err := authzed.NewClient(
		container.GetEndpoint(ctx),
		grpcutil.WithInsecureBearerToken(container.SecretKey()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	res, err := spicedbClient.SchemaServiceClient.WriteSchema(ctx, &v1.WriteSchemaRequest{
		Schema: testdata.MODEL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.WrittenAt == nil {
		t.Fatal("expected written_at to be set")
	}
}

func TestSpiceDBSecretCustomizer(t *testing.T) {
	ctx := context.Background()
	const secretKey = "testsecret"
	customizer := spicedbcontainer.SecretKeyCustomizer{
		SecretKey: secretKey,
	}
	container, err := spicedbcontainer.RunContainer(ctx,
		testcontainers.WithImage("authzed/spicedb:v1.33.0"),
		customizer,
	)

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	spicedbClient, err := authzed.NewClient(
		container.GetEndpoint(ctx),
		grpcutil.WithInsecureBearerToken(secretKey),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	res, err := spicedbClient.SchemaServiceClient.WriteSchema(ctx, &v1.WriteSchemaRequest{
		Schema: testdata.MODEL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.WrittenAt == nil {
		t.Fatal("expected written_at to be set")
	}
}

func TestSpiceModelCustomizer(t *testing.T) {
	ctx := context.Background()
	const defaultSecretKey = "somepresharedkey"
	modelCustomizer := spicedbcontainer.ModelCustomizer{
		SecretKey: defaultSecretKey,
		Model:     testdata.MODEL,
	}
	container, err := spicedbcontainer.RunContainer(ctx,
		testcontainers.WithImage("authzed/spicedb:v1.33.0"),
		modelCustomizer,
	)

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	spicedbClient, err := authzed.NewClient(
		container.GetEndpoint(ctx),
		grpcutil.WithInsecureBearerToken(defaultSecretKey),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	res, err := spicedbClient.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectId:   "testplatform",
					ObjectType: "platform",
				},
				Relation: "administrator",
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectId:   "testuser",
						ObjectType: "user",
					},
				},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.WrittenAt == nil {
		t.Fatal("expected written_at to be set")
	}
}
