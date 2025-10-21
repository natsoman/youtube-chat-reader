package domain

import "errors"

// Author represents a YouTube chat author
type Author struct {
	id              string
	name            string
	profileImageURL string
	isVerified      bool
}

func NewAuthor(id string, name string, profileImageURL string, isVerified bool) (*Author, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if profileImageURL == "" {
		return nil, errors.New("profile image url is empty")
	}

	return &Author{id: id, name: name, profileImageURL: profileImageURL, isVerified: isVerified}, nil
}

func (a *Author) ID() string {
	return a.id
}

func (a *Author) Name() string {
	return a.name
}

func (a *Author) ProfileImageURL() string {
	return a.profileImageURL
}

func (a *Author) IsVerified() bool {
	return a.isVerified
}
