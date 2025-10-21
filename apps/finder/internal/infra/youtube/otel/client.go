package otel

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

type Client interface {
	SearchUpcomingLiveStream(ctx context.Context, channelID string) ([]string, error)
	ListLiveStreams(ctx context.Context, videoIDs []string) ([]domain.LiveStream, error)
}

type InstrumentedClient struct {
	client Client
	tracer trace.Tracer
}

func NewInstrumentedClient(c Client) (*InstrumentedClient, error) {
	if c == nil {
		return nil, errors.New("client is nil")
	}

	return &InstrumentedClient{
		client: c,
		tracer: otel.Tracer("github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/youtube/otel"),
	}, nil
}

func (ic *InstrumentedClient) SearchUpcomingLiveStream(ctx context.Context, channelID string) ([]string, error) {
	spanCtx, span := ic.tracer.Start(ctx, "youtubeClient.searchUpcomingLiveStream")
	defer span.End()

	span.SetAttributes(attribute.String("channelId", channelID))

	upcomingLiveStreamIDs, err := ic.client.SearchUpcomingLiveStream(spanCtx, channelID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return upcomingLiveStreamIDs, nil
}

func (ic *InstrumentedClient) ListLiveStreams(ctx context.Context, videoIDs []string) ([]domain.LiveStream, error) {
	spanCtx, span := ic.tracer.Start(ctx, "youtubeClient.listLiveStreams")
	defer span.End()

	span.SetAttributes(attribute.StringSlice("videoIds", videoIDs))

	liveStreams, err := ic.client.ListLiveStreams(spanCtx, videoIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return liveStreams, nil
}
