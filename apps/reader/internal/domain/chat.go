package domain

// ChatMessages represents a batch of chat messages and their authors.
type ChatMessages struct {
	nextPageToken string
	textMessages  map[string]TextMessage
	bans          map[string]Ban
	donates       map[string]Donate
	authors       map[string]Author
}

func NewChatMessages(nextPageToken string) *ChatMessages {
	return &ChatMessages{
		nextPageToken: nextPageToken,
		textMessages:  make(map[string]TextMessage),
		bans:          make(map[string]Ban),
		donates:       make(map[string]Donate),
		authors:       make(map[string]Author),
	}
}

func (cm *ChatMessages) NextPageToken() string {
	return cm.nextPageToken
}

func (cm *ChatMessages) AddTextMessage(m *TextMessage) {
	if _, exists := cm.textMessages[m.ID()]; !exists {
		cm.textMessages[m.ID()] = *m
	}
}

func (cm *ChatMessages) TextMessages() []TextMessage {
	i := 0

	mm := make([]TextMessage, len(cm.textMessages))
	for _, m := range cm.textMessages {
		mm[i] = m
		i++
	}

	return mm
}

func (cm *ChatMessages) AddAuthor(a *Author) {
	if _, exists := cm.authors[a.ID()]; !exists {
		cm.authors[a.ID()] = *a
	}
}

func (cm *ChatMessages) Authors() []Author {
	i := 0

	aa := make([]Author, len(cm.authors))
	for _, a := range cm.authors {
		aa[i] = a
		i++
	}

	return aa
}

func (cm *ChatMessages) AddBan(b *Ban) {
	if _, exists := cm.bans[b.ID()]; !exists {
		cm.bans[b.ID()] = *b
	}
}

func (cm *ChatMessages) Bans() []Ban {
	i := 0

	bb := make([]Ban, len(cm.bans))
	for _, b := range cm.bans {
		bb[i] = b
		i++
	}

	return bb
}

func (cm *ChatMessages) AddDonate(d *Donate) {
	if _, exists := cm.donates[d.ID()]; !exists {
		cm.donates[d.ID()] = *d
	}
}

func (cm *ChatMessages) Donates() []Donate {
	i := 0

	dd := make([]Donate, len(cm.donates))
	for _, d := range cm.donates {
		dd[i] = d
		i++
	}

	return dd
}
