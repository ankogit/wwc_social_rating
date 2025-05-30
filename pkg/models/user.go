package models

import "time"

type User struct {
	ID             int64
	FirstName      string
	LastName       string
	UserName       string
	Score          int64
	IsLastUp       bool
	Achievements   string
	ProfileURL     string
	ScoreUpdatedAt *time.Time
}
