package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

type OutboxEvent struct {
	ID        string
	Topic     string
	Key       string
	Payload   []byte
	Published bool
}

type OutboxRepository interface {
	Pending(ctx context.Context) ([]OutboxEvent, error)
	MarkAsPublished(ctx context.Context, events []OutboxEvent) error
}

type OutboxSyncProducer struct {
	syncProducer sarama.SyncProducer
	outbox       OutboxRepository
}

func NewOutboxSyncProducer(syncProducer sarama.SyncProducer, outbox OutboxRepository) (*OutboxSyncProducer, error) {
	if syncProducer == nil {
		return nil, errors.New("sync producer is nil")
	}

	if outbox == nil {
		return nil, errors.New("outbox repository is nil")
	}

	return &OutboxSyncProducer{
		syncProducer: syncProducer,
		outbox:       outbox,
	}, nil
}

func (osp *OutboxSyncProducer) ProducePending(ctx context.Context) error {
	events, err := osp.outbox.Pending(ctx)
	if err != nil {
		return fmt.Errorf("pending events: %v", err)
	}

	if len(events) == 0 {
		return nil
	}

	pms := make([]*sarama.ProducerMessage, len(events))
	for i, e := range events {
		pms[i] = &sarama.ProducerMessage{
			Topic: e.Topic,
			Key:   sarama.StringEncoder(e.Key),
			Value: sarama.StringEncoder(e.Payload),
		}
	}

	if err = osp.syncProducer.SendMessages(pms); err != nil {
		return fmt.Errorf("send messages: %v", err)
	}

	if err = osp.outbox.MarkAsPublished(ctx, events); err != nil {
		return fmt.Errorf("mark event as published: %v", err)
	}

	return nil
}
