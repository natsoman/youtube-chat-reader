//go:generate mockgen -destination=mock_test.go -package=otel_test -source=find.go

package otel_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"

	infraotel "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/otel"
)

func TestInstrumentedLiveStreamFinder_Find(t *testing.T) {
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

			mockLiveStreamFinder := NewMockLiveStreamFinder(gomock.NewController(t))
			liveStreamFinder, err := infraotel.NewInstrumentedLiveStreamFinder(mockLiveStreamFinder)
			require.NotNil(t, liveStreamFinder)
			require.NoError(t, err)

			// Given
			mockLiveStreamFinder.EXPECT().
				Find(gomock.Any(), []string{"A", "B"}).
				Return(tc.expError)

			// When
			err = liveStreamFinder.Find(t.Context(), []string{"A", "B"})

			// Then
			assert.Equal(t, err, tc.expError)

			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
			}

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan(
				"liveStreamFinder.find",
				oteltrace.SpanKindInternal,
				status,
				attribute.StringSlice("channelIds", []string{"A", "B"}),
			)
		})
	}
}
