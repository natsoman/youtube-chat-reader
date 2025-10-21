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

type AuthorRepository interface {
	Upsert(ctx context.Context, aa []domain.Author) error
}

type InstrumentedAuthorRepository struct {
	repo   AuthorRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedAuthorRepository(repo AuthorRepository) (*InstrumentedAuthorRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("author repository is nil")
	}

	return &InstrumentedAuthorRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedAuthorRepository) Upsert(ctx context.Context, aa []domain.Author) error {
	spanCtx, span := r.tracer.Start(ctx, "authorRepository.upsert")
	defer span.End()

	if err := r.repo.Upsert(spanCtx, aa); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
