//nolint:dupl
package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type LiveStreamProgressRepository interface {
	Insert(ctx context.Context, lsp *domain.LiveStreamProgress) error
	Upsert(ctx context.Context, lsp *domain.LiveStreamProgress) error
	Started(ctx context.Context, startsWithin time.Duration) ([]domain.LiveStreamProgress, error)
}

type InstrumentedLiveStreamProgressRepository struct {
	repo   LiveStreamProgressRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedLiveStreamProgressRepository(repo LiveStreamProgressRepository) (
	*InstrumentedLiveStreamProgressRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("live stream progress repository is nil")
	}

	return &InstrumentedLiveStreamProgressRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedLiveStreamProgressRepository) Insert(ctx context.Context, lsp *domain.LiveStreamProgress) error {
	spanCtx, span := r.tracer.Start(ctx, "liveStreamProgressRepository.insert")
	defer span.End()

	if err := r.repo.Insert(spanCtx, lsp); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}

func (r *InstrumentedLiveStreamProgressRepository) Started(ctx context.Context, startsWithin time.Duration) (
	[]domain.LiveStreamProgress, error) {
	spanCtx, span := r.tracer.Start(ctx, "liveStreamProgressRepository.started")
	defer span.End()

	pp, err := r.repo.Started(spanCtx, startsWithin)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return pp, nil
}

func (r *InstrumentedLiveStreamProgressRepository) Upsert(ctx context.Context, lsp *domain.LiveStreamProgress) error {
	spanCtx, span := r.tracer.Start(ctx, "liveStreamProgressRepository.upsert")
	defer span.End()

	if err := r.repo.Upsert(spanCtx, lsp); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
