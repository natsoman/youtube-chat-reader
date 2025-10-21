//go:generate mockgen -destination=mock_test.go -package=otel_test -source=txn.go

package otel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	mongootel "github.com/natsoman/youtube-chat-reader/pkg/mongo/otel"
	"github.com/natsoman/youtube-chat-reader/pkg/otel/oteltest"
)

func TestInstrumentedTransactor_Atomic(t *testing.T) {
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

			mockTransactor := NewMockTransactor(gomock.NewController(t))
			instrumentedTransactor, err := mongootel.NewInstrumentedTransactor(mockTransactor)
			require.NotNil(t, instrumentedTransactor)
			require.NoError(t, err)

			// Given
			mockTransactor.EXPECT().
				Atomic(gomock.Any(), gomock.Any()).
				Return(tc.expError)

			// When
			err = instrumentedTransactor.Atomic(t.Context(), func(ctx context.Context) error { return nil })

			// Then
			assert.Equal(t, err, tc.expError)

			if tc.expError != nil {
				assert.EqualError(t, err, tc.expError.Error())
			}

			status := trace.Status{Code: tc.expStatusCode}
			if tc.expError != nil {
				status.Description = tc.expError.Error()
			}

			trc.AssertSpan("transactor.atomic", oteltrace.SpanKindInternal, status)
		})
	}
}
