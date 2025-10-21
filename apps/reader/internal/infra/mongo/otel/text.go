//nolint:dupl
package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type TextMessageRepository interface {
	Insert(ctx context.Context, tms []domain.TextMessage) error
}

type InstrumentedTextMessageRepository struct {
	repo   TextMessageRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedTextMessageRepository(repo TextMessageRepository) (*InstrumentedTextMessageRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("text message repository is nil")
	}

	return &InstrumentedTextMessageRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedTextMessageRepository) Insert(ctx context.Context, tms []domain.TextMessage) error {
	spanCtx, span := r.tracer.Start(ctx, "textMessageRepository.insert")
	defer span.End()

	if err := r.repo.Insert(spanCtx, tms); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
