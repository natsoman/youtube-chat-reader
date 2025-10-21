//go:generate mockgen -destination=mock_live_test.go -package=otel_test -source=live.go
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

	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/mongo/otel"
)

func TestInstrumentedLiveStreamRepository_Insert(t *testing.T) {
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
			instrumentedLiveStreamRepo, mockLiveStreamRepository := newMockInstrumentedLiveStreamRepo(t)

			now := time.Now()
			ls, err := domain.NewLiveStream("id1", "title1", "chanId1", "chanTitle1", "thumbUrl1", "chatId1", now, now)
			require.NoError(t, err)

			// Given
			mockLiveStreamRepository.EXPECT().
				Insert(gomock.Any(), []domain.LiveStream{*ls}).
				Return(tc.expError)

			// When
			err = instrumentedLiveStreamRepo.Insert(t.Context(), []domain.LiveStream{*ls})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("liveStreamRepository.insert", oteltrace.SpanKindInternal, status)
		})
	}
}

func TestInstrumentedLiveStreamRepository_Existing(t *testing.T) {
	testCases := []struct {
		name             string
		expLiveStreamIDs []string
		expError         error
		expStatusCode    codes.Code
	}{
		{
			name:             "ok",
			expLiveStreamIDs: []string{"A"},
			expStatusCode:    codes.Ok,
		},
		{
			name:          "error",
			expError:      errors.New("error"),
			expStatusCode: codes.Error,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trc := oteltest.NewTracer(t)

			instrumentedLiveStreamRepo, mockLiveStreamRepository := newMockInstrumentedLiveStreamRepo(t)

			// Given
			mockLiveStreamRepository.EXPECT().
				Existing(gomock.Any(), []string{"A"}).
				Return(tc.expLiveStreamIDs, tc.expError)

			// When
			actualLiveStreamIDs, err := instrumentedLiveStreamRepo.Existing(t.Context(), []string{"A"})

			// Then
			assert.Equal(t, tc.expLiveStreamIDs, actualLiveStreamIDs)
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("liveStreamRepository.existing", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedLiveStreamRepo(t *testing.T) (mongootel.LiveStreamRepository, *MockLiveStreamRepository) {
	t.Helper()

	mockLiveStreamRepository := NewMockLiveStreamRepository(gomock.NewController(t))
	instrumentedLiveStreamRepo, err := mongootel.NewInstrumentedLiveStreamRepository(mockLiveStreamRepository)
	require.NotNil(t, instrumentedLiveStreamRepo)
	require.NoError(t, err)

	return instrumentedLiveStreamRepo, mockLiveStreamRepository
}
