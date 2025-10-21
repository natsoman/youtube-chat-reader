package domain

import (
	"errors"
	"time"
)

type Donate struct {
	id           string
	authorID     string
	videoID      string
	comment      string
	amount       string
	amountMicros uint
	currency     string
	publishedAt  time.Time
}

func NewDonate(
	id string,
	authorID string,
	videoID string,
	comment string,
	amount string,
	amountMicros uint,
	currency string,
	publishedAt time.Time,
) (*Donate, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if authorID == "" {
		return nil, errors.New("author id is empty")
	}

	if videoID == "" {
		return nil, errors.New("video id is empty")
	}

	if amount == "" {
		return nil, errors.New("amount is empty")
	}

	if currency == "" {
		return nil, errors.New("currency is empty")
	}

	if publishedAt.IsZero() {
		return nil, errors.New("published at is zero")
	}

	return &Donate{
		id:           id,
		authorID:     authorID,
		videoID:      videoID,
		comment:      comment,
		amount:       amount,
		amountMicros: amountMicros,
		currency:     currency,
		publishedAt:  publishedAt,
	}, nil
}

func (d *Donate) ID() string {
	return d.id
}

func (d *Donate) AuthorID() string {
	return d.authorID
}

func (d *Donate) VideoID() string {
	return d.videoID
}

func (d *Donate) Comment() string {
	return d.comment
}

func (d *Donate) Amount() string {
	return d.amount
}

func (d *Donate) AmountMicros() uint {
	return d.amountMicros
}

func (d *Donate) Currency() string {
	return d.currency
}

func (d *Donate) PublishedAt() time.Time {
	return d.publishedAt
}
