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

type BanRepository interface {
	Insert(ctx context.Context, bb []domain.Ban) error
}

type InstrumentedBanRepository struct {
	repo   BanRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedBanRepository(repo BanRepository) (*InstrumentedBanRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("ban repository is nil")
	}

	return &InstrumentedBanRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedBanRepository) Insert(ctx context.Context, bb []domain.Ban) error {
	spanCtx, span := r.tracer.Start(ctx, "banRepository.insert")
	defer span.End()

	if err := r.repo.Insert(spanCtx, bb); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
