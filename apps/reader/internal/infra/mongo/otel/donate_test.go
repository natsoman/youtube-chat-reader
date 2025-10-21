//go:generate mockgen -destination=mock_donate_test.go -package=otel_test -source=donate.go
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

func TestInstrumentedDonateRepository_Insert(t *testing.T) {
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
			instrumentedDonateRepo, mockDonateRepository := newMockInstrumentedDonateRepo(t)

			d, err := domain.NewDonate("id", "authorId", "videoId", "comment", "amount", uint(10), "euro", time.Now())
			require.NoError(t, err)

			// Given
			mockDonateRepository.EXPECT().
				Insert(gomock.Any(), []domain.Donate{*d}).
				Return(tc.expError)

			// When
			err = instrumentedDonateRepo.Insert(t.Context(), []domain.Donate{*d})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("donateRepository.insert", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedDonateRepo(t *testing.T) (mongootel.DonateRepository, *MockDonateRepository) {
	t.Helper()

	mockDonateRepository := NewMockDonateRepository(gomock.NewController(t))
	instrumentedDonateRepo, err := mongootel.NewInstrumentedDonateRepository(mockDonateRepository)
	require.NotNil(t, instrumentedDonateRepo)
	require.NoError(t, err)

	return instrumentedDonateRepo, mockDonateRepository
}
