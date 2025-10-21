package domain

import (
	"errors"
	"time"
)

// LiveStream represents a YouTube video with live broadcast
type LiveStream struct {
	id             string
	channelID      string
	title          string
	thumbnailURL   string
	publishedAt    time.Time
	scheduledStart time.Time
	chatID         string
}

func NewLiveStream(id, title, channelID, channelTitle, thumbnailURL, chatID string,
	publishedAt, scheduledStart time.Time) (*LiveStream, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if title == "" {
		return nil, errors.New("title is empty")
	}

	if channelID == "" {
		return nil, errors.New("channel id is empty")
	}

	if channelTitle == "" {
		return nil, errors.New("channel title is empty")
	}

	if thumbnailURL == "" {
		return nil, errors.New("thumbnail URL is empty")
	}

	if chatID == "" {
		return nil, errors.New("chat id is empty")
	}

	if publishedAt.IsZero() {
		return nil, errors.New("published at is zero")
	}

	if scheduledStart.IsZero() {
		return nil, errors.New("scheduled start is zero")
	}

	return &LiveStream{
		id:             id,
		title:          title,
		channelID:      channelID,
		thumbnailURL:   thumbnailURL,
		publishedAt:    publishedAt,
		chatID:         chatID,
		scheduledStart: scheduledStart,
	}, nil
}

func (l LiveStream) ID() string {
	return l.id
}

func (l LiveStream) ChannelID() string {
	return l.channelID
}

func (l LiveStream) Title() string {
	return l.title
}

func (l LiveStream) ThumbnailURL() string {
	return l.thumbnailURL
}

func (l LiveStream) PublishedAt() time.Time {
	return l.publishedAt
}

func (l LiveStream) ChatID() string {
	return l.chatID
}

func (l LiveStream) ScheduledStart() time.Time {
	return l.scheduledStart
}
