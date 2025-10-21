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

type DonateRepository interface {
	Insert(ctx context.Context, dd []domain.Donate) error
}

type InstrumentedDonateRepository struct {
	repo   DonateRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedDonateRepository(repo DonateRepository) (*InstrumentedDonateRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("donate repository is nil")
	}

	return &InstrumentedDonateRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedDonateRepository) Insert(ctx context.Context, dd []domain.Donate) error {
	spanCtx, span := r.tracer.Start(ctx, "donateRepository.insert")
	defer span.End()

	if err := r.repo.Insert(spanCtx, dd); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
