package storage

import (
	"github.com/ankogit/wwc_social_rating/pkg/storage/stormDB"
	"github.com/asdine/storm/v3"
)

type Repositories struct {
	Chats ChatRepository
	Users UserRepository
	Polls PollRepository
}

func NewRepositories(db *storm.DB) *Repositories {
	return &Repositories{
		Chats: stormDB.NewChatRepository(db),
		Users: stormDB.NewUserRepository(db),
		Polls: stormDB.NewPollRepository(db),
	}
}
