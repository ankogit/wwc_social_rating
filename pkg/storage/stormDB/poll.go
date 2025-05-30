package stormDB

import (
	"github.com/ankogit/wwc_social_rating/pkg/models"
	"github.com/asdine/storm/v3"
)

type PollRepository struct {
	db *storm.DB
}

func NewPollRepository(db *storm.DB) *PollRepository {
	return &PollRepository{db: db}
}

func (r *PollRepository) Save(data models.Poll) error {
	return r.db.Save(&data)
}

func (r *PollRepository) Get(pollId int64) (poll models.Poll, err error) {
	err = r.db.One("ID", pollId, &poll)
	return
}
func (r *PollRepository) Delete(poll models.Poll) error {
	return r.db.DeleteStruct(&poll)
}

func (r *PollRepository) All() (polls []models.Poll, err error) {
	err = r.db.All(&polls)
	return
}

func (r *PollRepository) AllActive() (polls []models.Poll, err error) {
	err = r.db.Find("IsClosed", false, &polls)
	return
}
