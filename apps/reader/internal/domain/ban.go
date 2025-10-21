package domain

import (
	"errors"
	"fmt"
	"time"
)

const (
	Permanent BanType = iota + 1
	Temporary
)

type BanType int

func (bt BanType) String() string {
	switch bt {
	case Temporary:
		return "temporary"
	case Permanent:
		return "permanent"
	}

	return ""
}

// Ban represents a YouTube ban
type Ban struct {
	id          string
	authorID    string
	videoID     string
	banType     BanType
	duration    time.Duration
	publishedAt time.Time
}

func NewBan(id, authorID, videoID, banType string, duration time.Duration, publishedAt time.Time) (*Ban, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if authorID == "" {
		return nil, errors.New("author id is empty")
	}

	if videoID == "" {
		return nil, errors.New("video id is empty")
	}

	b := &Ban{
		id:          id,
		authorID:    authorID,
		videoID:     videoID,
		duration:    duration,
		publishedAt: publishedAt,
	}

	switch banType {
	case "TEMPORARY", "temporary":
		if duration == 0 {
			return nil, errors.New("duration is zero")
		}

		b.banType = Temporary
		b.duration = duration
	case "PERMANENT", "permanent":
		b.banType = Permanent
	default:
		return nil, fmt.Errorf("unknown ban type '%s'", banType)
	}

	return b, nil
}

func (b *Ban) ID() string {
	return b.id
}

func (b *Ban) AuthorID() string {
	return b.authorID
}

func (b *Ban) VideoID() string {
	return b.videoID
}

func (b *Ban) BanType() BanType {
	return b.banType
}

func (b *Ban) Duration() time.Duration {
	return b.duration
}

func (b *Ban) PublishedAt() time.Time {
	return b.publishedAt
}
