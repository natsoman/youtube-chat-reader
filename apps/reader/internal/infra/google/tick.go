package google

import (
	"time"
)

type Ticker struct{}

func (t *Ticker) Start(d time.Duration) (<-chan time.Time, func()) {
	ticker := time.NewTicker(d)

	return ticker.C, func() { ticker.Stop() }
}
