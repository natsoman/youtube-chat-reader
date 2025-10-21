//nolint:dupl
//go:generate mockgen -destination=mock_progress_test.go -package=otel_test -source=progress.go
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

func TestInstrumentedLiveStreamProgressRepository_Insert(t *testing.T) {
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
			instrumentedLiveStreamProgressRepo, mockLiveStreamProgressRepository :=
				newMockInstrumentedLiveStreamProgressRepo(t)

			lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
			require.NoError(t, err)

			// Given
			mockLiveStreamProgressRepository.EXPECT().
				Insert(gomock.Any(), lsp).
				Return(tc.expError)

			// When
			err = instrumentedLiveStreamProgressRepo.Insert(t.Context(), lsp)

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("liveStreamProgressRepository.insert", oteltrace.SpanKindInternal, status)
		})
	}
}

func TestInstrumentedLiveStreamProgressRepository_Started(t *testing.T) {
	testCases := []struct {
		name                   string
		expError               error
		expLiveStreamsProgress func(t *testing.T) []domain.LiveStreamProgress
		expStatusCode          codes.Code
	}{
		{
			name: "ok",
			expLiveStreamsProgress: func(t *testing.T) []domain.LiveStreamProgress {
				lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
				require.NoError(t, err)

				return []domain.LiveStreamProgress{*lsp}
			},
			expStatusCode: codes.Ok,
		},
		{
			name: "error",
			expLiveStreamsProgress: func(t *testing.T) []domain.LiveStreamProgress {
				return nil
			},
			expStatusCode: codes.Error,
			expError:      errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trc := oteltest.NewTracer(t)
			instrumentedLiveStreamProgressRepo, mockLiveStreamProgressRepository :=
				newMockInstrumentedLiveStreamProgressRepo(t)

			expLiveStreamsProgress := tc.expLiveStreamsProgress(t)

			// Given
			mockLiveStreamProgressRepository.EXPECT().
				Started(gomock.Any(), time.Hour).
				Return(expLiveStreamsProgress, tc.expError)

			// When
			actLiveStreamsProgress, err := instrumentedLiveStreamProgressRepo.Started(t.Context(), time.Hour)

			// Then
			assert.Equal(t, err, tc.expError)
			assert.Equal(t, expLiveStreamsProgress, actLiveStreamsProgress)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("liveStreamProgressRepository.started", oteltrace.SpanKindInternal, status)
		})
	}
}

func TestInstrumentedLiveStreamProgressRepository_Upsert(t *testing.T) {
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
			instrumentedLiveStreamProgressRepo, mockLiveStreamProgressRepository :=
				newMockInstrumentedLiveStreamProgressRepo(t)

			lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
			require.NoError(t, err)

			// Given
			mockLiveStreamProgressRepository.EXPECT().
				Upsert(gomock.Any(), lsp).
				Return(tc.expError)

			// When
			err = instrumentedLiveStreamProgressRepo.Upsert(t.Context(), lsp)

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("liveStreamProgressRepository.upsert", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedLiveStreamProgressRepo(t *testing.T) (
	mongootel.LiveStreamProgressRepository, *MockLiveStreamProgressRepository) {
	t.Helper()

	mockLiveStreamProgressRepository := NewMockLiveStreamProgressRepository(gomock.NewController(t))
	instrumentedLiveStreamProgressRepo, err :=
		mongootel.NewInstrumentedLiveStreamProgressRepository(mockLiveStreamProgressRepository)
	require.NotNil(t, instrumentedLiveStreamProgressRepo)
	require.NoError(t, err)

	return instrumentedLiveStreamProgressRepo, mockLiveStreamProgressRepository
}
