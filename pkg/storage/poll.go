package storage

import "github.com/ankogit/wwc_social_rating/pkg/models"

type PollRepository interface {
	Save(data models.Poll) error
	Get(pollId int64) (models.Poll, error)
	Delete(poll models.Poll) error
	All() ([]models.Poll, error)
	AllActive() ([]models.Poll, error)
}
