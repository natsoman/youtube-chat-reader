package domain

import (
	"errors"
	"time"
)

// LiveStreamProgress represents a YouTube live stream reading progress.
type LiveStreamProgress struct {
	// id contains the identifier of the live stream.
	id string
	// chatID contains the identifier of the chat id of the live stream.
	chatID string
	// nextPageToken contains the next page token that should be used to fetch the next page of messages. If empty,
	// the reading of the live stream has not been started or has been finished without any message.
	nextPageToken string
	// scheduledStart indicates the scheduled start time of the live stream.
	scheduledStart time.Time
	// finishedAt indicate when the reading of the live stream has been finished. If nil, the reading of
	// the live stream has not been started or is in progress.
	finishedAt *time.Time
	// finishReason describes why the reading has been finished
	finishReason string
}

func NewLiveStreamProgress(id, chatID string, scheduledStart time.Time) (*LiveStreamProgress, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if chatID == "" {
		return nil, errors.New("chat id is empty")
	}

	if scheduledStart.IsZero() {
		return nil, errors.New("scheduled start is zero")
	}

	return &LiveStreamProgress{
		id:             id,
		chatID:         chatID,
		scheduledStart: scheduledStart,
	}, nil
}

// ID returns the identifier of the live stream.
func (lsp *LiveStreamProgress) ID() string {
	return lsp.id
}

// ChatID returns the identifier of the chat id of the live stream.
func (lsp *LiveStreamProgress) ChatID() string {
	return lsp.chatID
}

// NextPageToken returns the next page token for fetching messages.
func (lsp *LiveStreamProgress) NextPageToken() string {
	return lsp.nextPageToken
}

// SetNextPageToken sets the next page token.
func (lsp *LiveStreamProgress) SetNextPageToken(token string) {
	lsp.nextPageToken = token
}

// ScheduledStart returns the scheduled start time of the live stream.
func (lsp *LiveStreamProgress) ScheduledStart() time.Time {
	return lsp.scheduledStart
}

// FinishedAt returns when the reading was finished, or nil if still in progress.
func (lsp *LiveStreamProgress) FinishedAt() *time.Time {
	return lsp.finishedAt
}

// FinishReason returns the reason why the reading was finished.
func (lsp *LiveStreamProgress) FinishReason() string {
	return lsp.finishReason
}

// Finish marks the live stream as finished with the given reason.
func (lsp *LiveStreamProgress) Finish(at time.Time, reason string) {
	lsp.finishedAt = &at
	lsp.finishReason = reason
}

// IsFinished indicates if the live stream progress has been finished.
func (lsp *LiveStreamProgress) IsFinished() bool {
	return lsp.finishedAt != nil
}
