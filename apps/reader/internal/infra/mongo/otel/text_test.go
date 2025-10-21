//go:generate mockgen -destination=mock_text_test.go -package=otel_test -source=text.go
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

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo/otel"
	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"
)

func TestInstrumentedTextMessageRepository_Insert(t *testing.T) {
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
			instrumentedTextMessageRepo, mockTextMessageRepository := newMockInstrumentedTextMessageRepo(t)

			tm, err := domain.NewTextMessage("id", "videoId", "authorId", "text", time.Now())
			require.NoError(t, err)

			// Given
			mockTextMessageRepository.EXPECT().
				Insert(gomock.Any(), []domain.TextMessage{*tm}).
				Return(tc.expError)

			// When
			err = instrumentedTextMessageRepo.Insert(t.Context(), []domain.TextMessage{*tm})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("textMessageRepository.insert", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedTextMessageRepo(t *testing.T) (mongootel.TextMessageRepository, *MockTextMessageRepository) {
	t.Helper()

	mockTextMessageRepository := NewMockTextMessageRepository(gomock.NewController(t))
	instrumentedTextMessageRepo, err := mongootel.NewInstrumentedTextMessageRepository(mockTextMessageRepository)
	require.NotNil(t, instrumentedTextMessageRepo)
	require.NoError(t, err)

	return instrumentedTextMessageRepo, mockTextMessageRepository
}
