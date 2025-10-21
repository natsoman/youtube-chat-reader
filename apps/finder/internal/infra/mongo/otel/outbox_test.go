//go:generate mockgen -destination=mock_outbox_test.go -package=otel_test -source=outbox.go
//nolint:dupl
package otel_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/pkg/kafka"
	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/mongo/otel"
)

func TestInstrumentedOutboxRepository_InsertLiveStreamsFound(t *testing.T) {
	testCases := []struct {
		name          string
		expError      error
		expStatusCode codes.Code
	}{
		{
			name:          "ok",
			expStatusCode: codes.Ok,
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
			instrumentedOutboxRepo, mockOutboxRepository := newMockInstrumentedOutboxRepo(t)

			now := time.Now()
			ls, err := domain.NewLiveStream("id1", "title1", "chanId1", "chanTitle1", "thumbUrl1", "chatId1", now, now)
			require.NoError(t, err)

			// Given
			mockOutboxRepository.EXPECT().
				InsertLiveStreamsFound(gomock.Any(), []domain.LiveStream{*ls}).
				Return(tc.expError)

			// When
			err = instrumentedOutboxRepo.InsertLiveStreamsFound(t.Context(), []domain.LiveStream{*ls})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("outboxRepository.insertLiveStreamsFound", oteltrace.SpanKindInternal, status)
		})
	}
}

func TestInstrumentedOutboxRepository_Pending(t *testing.T) {
	testCases := []struct {
		name            string
		expOutboxEvents []kafka.OutboxEvent
		expError        error
		expStatusCode   codes.Code
	}{
		{
			name:            "ok",
			expOutboxEvents: []kafka.OutboxEvent{{ID: "a"}, {ID: "b"}},
			expStatusCode:   codes.Ok,
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
			instrumentedOutboxRepo, mockOutboxRepository := newMockInstrumentedOutboxRepo(t)

			// Given
			mockOutboxRepository.EXPECT().
				Pending(gomock.Any()).
				Return(tc.expOutboxEvents, tc.expError)

			// When
			actOutboxEvents, err := instrumentedOutboxRepo.Pending(t.Context())

			// Then
			assert.Equal(t, tc.expOutboxEvents, actOutboxEvents)
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("outboxRepository.pending", oteltrace.SpanKindInternal, status)
		})
	}
}

func TestInstrumentedOutboxRepository_MarkAsPublished(t *testing.T) {
	testCases := []struct {
		name          string
		expError      error
		expStatusCode codes.Code
	}{
		{
			name:          "ok",
			expStatusCode: codes.Ok,
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
			instrumentedOutboxRepo, mockOutboxRepository := newMockInstrumentedOutboxRepo(t)

			// Given
			mockOutboxRepository.EXPECT().
				MarkAsPublished(gomock.Any(), []kafka.OutboxEvent{{ID: "a"}}).
				Return(tc.expError)

			// When
			err := instrumentedOutboxRepo.MarkAsPublished(t.Context(), []kafka.OutboxEvent{{ID: "a"}})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("outboxRepository.markAsPublished", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedOutboxRepo(t *testing.T) (mongootel.OutboxRepository, *MockOutboxRepository) {
	t.Helper()

	mockOutboxRepository := NewMockOutboxRepository(gomock.NewController(t))
	instrumentedOutboxRepo, err := mongootel.NewInstrumentedOutboxRepository(mockOutboxRepository)
	require.NotNil(t, instrumentedOutboxRepo)
	require.NoError(t, err)

	return instrumentedOutboxRepo, mockOutboxRepository
}
