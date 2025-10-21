//go:generate mockgen -destination=mock_test.go -package=otel_test -source=client.go
//nolint:dupl
package otel_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/youtube/otel"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

func TestInstrumentedClient_SearchUpcomingLiveStream(t *testing.T) {
	testCases := []struct {
		name                     string
		expUpcomingLiveStreamIDs []string
		expError                 error
		expStatusCode            codes.Code
	}{
		{
			name:                     "ok",
			expUpcomingLiveStreamIDs: []string{"liveStreamId"},
			expStatusCode:            codes.Ok,
		},
		{
			name:          "error",
			expStatusCode: codes.Error,
			expError:      errors.New("error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trc := oteltest.NewTracer(t)
			instrumentedClient, mockClient := newMockClient(t)

			// Given
			mockClient.EXPECT().
				SearchUpcomingLiveStream(gomock.Any(), "chanId").
				Return(tc.expUpcomingLiveStreamIDs, tc.expError)

			// When
			actUpcomingLiveStreamIDs, err := instrumentedClient.SearchUpcomingLiveStream(t.Context(), "chanId")

			// Then
			assert.Equal(t, tc.expUpcomingLiveStreamIDs, actUpcomingLiveStreamIDs)
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan(
				"youtubeClient.searchUpcomingLiveStream",
				oteltrace.SpanKindInternal,
				status,
				attribute.String("channelId", "chanId"),
			)
		})
	}
}

func TestInstrumentedClient_ListLiveStreams(t *testing.T) {
	now := time.Now()
	ls, err := domain.NewLiveStream("id1", "title1", "channelId1", "chanTitle1", "thumbUrl1", "chatId1", now, now)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		expLiveStreams []domain.LiveStream
		expError       error
		expStatusCode  codes.Code
	}{
		{
			name:           "ok",
			expLiveStreams: []domain.LiveStream{*ls},
			expStatusCode:  codes.Ok,
		},
		{
			name:          "error",
			expStatusCode: codes.Error,
			expError:      errors.New("error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trc := oteltest.NewTracer(t)
			instrumentedClient, mockClient := newMockClient(t)

			// Given
			mockClient.EXPECT().
				ListLiveStreams(gomock.Any(), []string{"id"}).
				Return(tc.expLiveStreams, tc.expError)

			// When
			actLiveStreams, err := instrumentedClient.ListLiveStreams(t.Context(), []string{"id"})

			// Then
			assert.Equal(t, tc.expLiveStreams, actLiveStreams)
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan(
				"youtubeClient.listLiveStreams",
				oteltrace.SpanKindInternal,
				status,
				attribute.StringSlice("videoIds", []string{"id"}),
			)
		})
	}
}

func newMockClient(t *testing.T) (*otel.InstrumentedClient, *MockClient) {
	t.Helper()

	mockClient := NewMockClient(gomock.NewController(t))
	instrumentedClient, err := otel.NewInstrumentedClient(mockClient)
	require.NotNil(t, instrumentedClient)
	require.NoError(t, err)

	return instrumentedClient, mockClient
}
