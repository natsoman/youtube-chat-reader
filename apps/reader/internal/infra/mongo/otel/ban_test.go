//go:generate mockgen -destination=mock_ban_test.go -package=otel_test -source=ban.go
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

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo/otel"
	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"
)

func TestInstrumentedBanRepository_Insert(t *testing.T) {
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
			instrumentedBanRepo, mockBanRepository := newMockInstrumentedBanRepo(t)

			b, err := domain.NewBan("id", "authorId", "videoId", domain.Permanent.String(), 0, time.Now())
			require.NoError(t, err)

			// Given
			mockBanRepository.EXPECT().
				Insert(gomock.Any(), []domain.Ban{*b}).
				Return(tc.expError)

			// When
			err = instrumentedBanRepo.Insert(t.Context(), []domain.Ban{*b})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("banRepository.insert", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedBanRepo(t *testing.T) (mongootel.BanRepository, *MockBanRepository) {
	t.Helper()

	mockBanRepository := NewMockBanRepository(gomock.NewController(t))
	instrumentedBanRepo, err := mongootel.NewInstrumentedBanRepository(mockBanRepository)
	require.NotNil(t, instrumentedBanRepo)
	require.NoError(t, err)

	return instrumentedBanRepo, mockBanRepository
}
