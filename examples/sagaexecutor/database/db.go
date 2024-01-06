package database

import (
	"context"
	"fmt"
	"log"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/bindings/postgres"
	"github.com/dapr/components-contrib/metadata"
	"github.com/dapr/kit/logger"
)

func OpenDBConnection(connectionString string) (*postgres.Postgres, *bindings.Metadata) {

	// live DB test
	bPg := postgres.NewPostgres(logger.NewLogger("test")).(*postgres.Postgres)
	mPg := bindings.Metadata{Base: metadata.Base{Properties: map[string]string{"connectionString": connectionString}}}
	if err := bPg.Init(context.Background(), mPg); err != nil {
		log.Fatalf("Unable to connect to database: %s\n", err)
	}

	return bPg, &mPg
}

func CloseDBConnection(ctx context.Context, bPg *postgres.Postgres) {
	req := &bindings.InvokeRequest{
		Operation: "exec",
		Metadata:  map[string]string{},
	}
	req.Operation = "close"
	req.Metadata = nil
	req.Data = nil
	_, err := bPg.Invoke(ctx, req)
	if err != nil {
		fmt.Errorf("Error on DB close %s", err)
	}

	err = bPg.Close()
	if err != nil {
		log.Fatalln("Error on binding close", err)
	}

}
