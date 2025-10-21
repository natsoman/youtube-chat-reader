package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type LiveStreamProgressRepository interface {
	// Insert adds the provided live stream progress to the repository, ignoring duplicates.
	Insert(ctx context.Context, lsp *domain.LiveStreamProgress) error
}

type liveStreamFoundEventPayload struct {
	VideoID        string    `json:"videoId"`
	ChatID         string    `json:"chatId"`
	ScheduledStart time.Time `json:"scheduledStart"`
}

type LiveStreamFoundEventHandler struct {
	lsr LiveStreamProgressRepository
}

func NewLiveStreamFoundEventHandler(lsr LiveStreamProgressRepository) (*LiveStreamFoundEventHandler, error) {
	if lsr == nil {
		return nil, errors.New("live stream repository is nil")
	}

	return &LiveStreamFoundEventHandler{lsr: lsr}, nil
}

func (h *LiveStreamFoundEventHandler) Handle(ctx context.Context, m *sarama.ConsumerMessage) error {
	timeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var p liveStreamFoundEventPayload
	if err := json.Unmarshal(m.Value, &p); err != nil {
		return fmt.Errorf("unmarshal event payload: %v", err)
	}

	lsp, err := domain.NewLiveStreamProgress(p.VideoID, p.ChatID, p.ScheduledStart)
	if err != nil {
		return fmt.Errorf("new live stream: %v", err)
	}

	if err = h.lsr.Insert(timeCtx, lsp); err != nil {
		return fmt.Errorf("insert live stream: %v", err)
	}

	return nil
}
