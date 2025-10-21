//go:generate mockgen -destination=mock_author_test.go -package=otel_test -source=author.go
//nolint:dupl
package otel_test

import (
	"errors"
	"testing"

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

func TestInstrumentedAuthorRepository_Upsert(t *testing.T) {
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
			instrumentedAuthorRepo, mockAuthorRepository := newMockInstrumentedAuthorRepo(t)

			a, err := domain.NewAuthor("id", "name", "image", true)
			require.NoError(t, err)

			// Given
			mockAuthorRepository.EXPECT().
				Upsert(gomock.Any(), []domain.Author{*a}).
				Return(tc.expError)

			// When
			err = instrumentedAuthorRepo.Upsert(t.Context(), []domain.Author{*a})

			// Then
			assert.Equal(t, err, tc.expError)

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("authorRepository.upsert", oteltrace.SpanKindInternal, status)
		})
	}
}

func newMockInstrumentedAuthorRepo(t *testing.T) (mongootel.AuthorRepository, *MockAuthorRepository) {
	t.Helper()

	mockAuthorRepository := NewMockAuthorRepository(gomock.NewController(t))
	instrumentedAuthorRepo, err := mongootel.NewInstrumentedAuthorRepository(mockAuthorRepository)
	require.NotNil(t, instrumentedAuthorRepo)
	require.NoError(t, err)

	return instrumentedAuthorRepo, mockAuthorRepository
}
