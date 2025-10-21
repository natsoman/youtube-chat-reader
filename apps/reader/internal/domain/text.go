package domain

import (
	"errors"
	"time"
)

// TextMessage represents a YouTube text message
type TextMessage struct {
	id          string
	videoID     string
	authorID    string
	text        string
	publishedAt time.Time
}

func NewTextMessage(id, videoID, authorID, text string, publishedAt time.Time) (*TextMessage, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if videoID == "" {
		return nil, errors.New("video id is empty")
	}

	if authorID == "" {
		return nil, errors.New("author id is empty")
	}

	if publishedAt.IsZero() {
		return nil, errors.New("published at is zero")
	}

	return &TextMessage{
		id:          id,
		videoID:     videoID,
		authorID:    authorID,
		text:        text,
		publishedAt: publishedAt,
	}, nil
}

func (tm *TextMessage) ID() string {
	return tm.id
}

func (tm *TextMessage) VideoID() string {
	return tm.videoID
}

func (tm *TextMessage) AuthorID() string {
	return tm.authorID
}

func (tm *TextMessage) Text() string {
	return tm.text
}

func (tm *TextMessage) PublishedAt() time.Time {
	return tm.publishedAt
}
