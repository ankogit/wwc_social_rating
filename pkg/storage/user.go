package storage

import "github.com/ankogit/wwc_social_rating/pkg/models"

type UserRepository interface {
	Save(data models.User) error
	Get(chatId int64) (models.User, error)
	Delete(chat models.User) error
	All() ([]models.User, error)
}
