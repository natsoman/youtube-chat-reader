package domain

import "errors"

var (
	ErrChatNotFound          = errors.New("chat not found")
	ErrChatOffline           = errors.New("chat is offline")
	ErrUnavailableLiveStream = errors.New("unavailable live stream")
)
