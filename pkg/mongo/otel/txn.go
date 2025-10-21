package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Transactor interface {
	Atomic(ctx context.Context, fn func(ctx context.Context) error) error
}

type InstrumentedTransactor struct {
	txn    Transactor
	tracer oteltrace.Tracer
}

func NewInstrumentedTransactor(txn Transactor) (*InstrumentedTransactor, error) {
	if txn == nil {
		return nil, fmt.Errorf("transactor is nil")
	}

	return &InstrumentedTransactor{
		txn:    txn,
		tracer: otel.Tracer("github.com/natsoman/youtube-chat-reader/pkg/mongo/otel"),
	}, nil
}

func (it *InstrumentedTransactor) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
	spanCtx, span := it.tracer.Start(ctx, "transactor.atomic")
	defer span.End()

	if err := it.txn.Atomic(spanCtx, fn); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
