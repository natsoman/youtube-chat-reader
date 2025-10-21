package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

var _ sarama.ConsumerGroupHandler = (*ConsumerGroupHandler)(nil)

type MessageHandler func(ctx context.Context, m *sarama.ConsumerMessage) error

type ConsumerGroupHandler struct {
	log         *slog.Logger
	msgHandlers map[string]MessageHandler
	// timeout of message processing. Must be less than Config.Consumer.Group.Session.Timeout
	timeout time.Duration
}

func NewConsumerGroupHandler(l *slog.Logger, msgHandlers map[string]MessageHandler, timeout time.Duration) (
	*ConsumerGroupHandler, error) {
	if l == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	const defaultTimeout = 10 * time.Second
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &ConsumerGroupHandler{
		log:         l,
		msgHandlers: msgHandlers,
		timeout:     timeout,
	}, nil
}

func (cgh *ConsumerGroupHandler) ConsumeClaim(cgs sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for cm := range claim.Messages() {
		if err := cgh.process(cgs, cm); err != nil {
			return err
		}
	}

	return nil
}

func (cgh *ConsumerGroupHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	return nil
}

func (cgh *ConsumerGroupHandler) Setup(s sarama.ConsumerGroupSession) error {
	return nil
}

func (cgh *ConsumerGroupHandler) process(cgs sarama.ConsumerGroupSession, cm *sarama.ConsumerMessage) error {
	timeCtx, cancel := context.WithTimeout(cgs.Context(), cgh.timeout)
	defer cancel()

	log := cgh.log.With(
		"topic", cm.Topic,
		"val", string(cm.Value),
		"key", cm.Key,
		"partition", cm.Partition,
		"offset", cm.Offset,
	)

	handler, exists := cgh.msgHandlers[cm.Topic]
	if !exists {
		log.WarnContext(cgs.Context(), "No registered handler", "topic", cm.Topic)
		cgs.MarkMessage(cm, "")

		return nil
	}

	if err := handler(timeCtx, cm); err != nil {
		// TODO implement a Dead Letter Queue
		log.ErrorContext(timeCtx, "Failed to handle message", "err", err)

		return fmt.Errorf("handle message: %v", err)
	}

	cgs.MarkMessage(cm, "")
	log.DebugContext(timeCtx, "Message handled")

	return nil
}
