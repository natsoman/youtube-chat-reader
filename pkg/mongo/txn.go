package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Transactor struct {
	client *mongo.Client
}

func NewTransactor(client *mongo.Client) (*Transactor, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	return &Transactor{client: client}, nil
}

func (m *Transactor) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(ctx)

	wrapFn := func(sessCtx mongo.SessionContext) (any, error) {
		return nil, fn(sessCtx)
	}

	opts := options.Transaction().
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Snapshot())

	_, err = session.WithTransaction(ctx, wrapFn, opts)
	if err != nil {
		return err
	}

	return nil
}
