package app

import (
	"errors"
	"time"
)

type Option func(*LiveStreamReader) error

// WithRetryInterval specifies the maximum interval between live stream reading attempts.
func WithRetryInterval(d time.Duration) Option {
	return func(s *LiveStreamReader) error {
		if d >= time.Second*10 && d <= time.Minute {
			s.retryInterval = d

			return nil
		}

		return errors.New("max retry interval must be gte 10sec and lte a minute")
	}
}

// // WithAdvanceStart specifies how much earlier the live stream reading should begin before the scheduled start time.
func WithAdvanceStart(d time.Duration) Option {
	return func(s *LiveStreamReader) error {
		if d >= time.Minute && d <= time.Hour {
			s.advanceStart = d

			return nil
		}

		return errors.New("starts within must be gte a minute and lte an hour")
	}
}
