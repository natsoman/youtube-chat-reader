package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

type LiveStreamRepository interface {
	Insert(ctx context.Context, liveStreams []domain.LiveStream) error
	Existing(ctx context.Context, liveStreamIDs []string) ([]string, error)
}

type InstrumentedLiveStreamRepository struct {
	repo   LiveStreamRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedLiveStreamRepository(repo LiveStreamRepository) (*InstrumentedLiveStreamRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("live stream repository is nil")
	}

	return &InstrumentedLiveStreamRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedLiveStreamRepository) Insert(ctx context.Context, liveStreams []domain.LiveStream) error {
	spanCtx, span := r.tracer.Start(ctx, "liveStreamRepository.insert")
	defer span.End()

	if err := r.repo.Insert(spanCtx, liveStreams); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}

func (r *InstrumentedLiveStreamRepository) Existing(ctx context.Context, liveStreamIDs []string) ([]string, error) {
	spanCtx, span := r.tracer.Start(ctx, "liveStreamRepository.existing")
	defer span.End()

	existingLiveStreamIDs, err := r.repo.Existing(spanCtx, liveStreamIDs)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return existingLiveStreamIDs, nil
}
