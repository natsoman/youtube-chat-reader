package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/natsoman/youtube-chat-reader/pkg/kafka"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

type OutboxRepository interface {
	InsertLiveStreamsFound(ctx context.Context, ls []domain.LiveStream) error
	Pending(ctx context.Context) ([]kafka.OutboxEvent, error)
	MarkAsPublished(ctx context.Context, events []kafka.OutboxEvent) error
}

type InstrumentedOutboxRepository struct {
	repo   OutboxRepository
	tracer oteltrace.Tracer
}

func NewInstrumentedOutboxRepository(repo OutboxRepository) (*InstrumentedOutboxRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("outbox repository is nil")
	}

	return &InstrumentedOutboxRepository{
		repo:   repo,
		tracer: otel.Tracer(pkgName),
	}, nil
}

func (r *InstrumentedOutboxRepository) InsertLiveStreamsFound(ctx context.Context, ls []domain.LiveStream) error {
	spanCtx, span := r.tracer.Start(ctx, "outboxRepository.insertLiveStreamsFound")
	defer span.End()

	if err := r.repo.InsertLiveStreamsFound(spanCtx, ls); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}

func (r *InstrumentedOutboxRepository) Pending(ctx context.Context) ([]kafka.OutboxEvent, error) {
	spanCtx, span := r.tracer.Start(ctx, "outboxRepository.pending")
	defer span.End()

	events, err := r.repo.Pending(spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return events, nil
}

func (r *InstrumentedOutboxRepository) MarkAsPublished(ctx context.Context, events []kafka.OutboxEvent) error {
	spanCtx, span := r.tracer.Start(ctx, "outboxRepository.markAsPublished")
	defer span.End()

	if err := r.repo.MarkAsPublished(spanCtx, events); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
