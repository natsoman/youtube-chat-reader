package otel

import (
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

type LiveStreamFinder interface {
	Find(ctx context.Context, channelIDs []string) error
}

type InstrumentedLiveStreamFinder struct {
	liveStreamFinder LiveStreamFinder
	tracer           oteltrace.Tracer
}

func NewInstrumentedLiveStreamFinder(liveStreamFinder LiveStreamFinder) (*InstrumentedLiveStreamFinder, error) {
	if liveStreamFinder == nil {
		return nil, errors.New("live stream finder is nil")
	}

	return &InstrumentedLiveStreamFinder{
		liveStreamFinder: liveStreamFinder,
		tracer:           otel.Tracer("github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/otel"),
	}, nil
}

func (f *InstrumentedLiveStreamFinder) Find(ctx context.Context, channelIDs []string) error {
	spanCtx, span := f.tracer.Start(ctx, "liveStreamFinder.find")
	defer span.End()

	span.SetAttributes(attribute.StringSlice("channelIds", channelIDs))

	if err := f.liveStreamFinder.Find(spanCtx, channelIDs); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
