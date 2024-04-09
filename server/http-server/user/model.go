package user

import (
	"time"

	"github.com/gocql/gocql"
)

type User struct {
	UserID    gocql.UUID `json:"user_id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	CreatedAt time.Time  `json:"created_at"`
}
