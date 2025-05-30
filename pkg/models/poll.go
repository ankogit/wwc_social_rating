package models

import (
	"time"
)

type Poll struct {
	ID        int
	PollID    string
	UserID    int64 //наш
	IsClosed  bool
	CreatedAt time.Time
}
