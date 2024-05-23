package stormDB

import (
	"github.com/ankogit/wwc_social_rating/pkg/models"
	"github.com/asdine/storm/v3"
)

type UserRepository struct {
	db *storm.DB
}

func NewUserRepository(db *storm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(data models.User) error {
	return r.db.Save(&data)
}

func (r *UserRepository) Get(id int64) (models.User, error) {
	var user models.User
	err := r.db.One("ID", id, &user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (user models.User, err error) {
	if len(username) > 1 && username[0] == '@' {
		username = username[1:]
	}
	err = r.db.One("UserName", username, &user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) Delete(user models.User) error {
	return r.db.DeleteStruct(&user)
}

func (r *UserRepository) All() ([]models.User, error) {
	var users []models.User
	err := r.db.All(&users)
	if err != nil {
		return users, err
	}
	return users, nil
}
