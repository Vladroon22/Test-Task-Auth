package database

import (
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/internal/sessions"
)

type Handlers interface {
	GetToken(id int) (string, error)
	SaveSession(id int, email, ip string, regTime time.Time, expireTime time.Duration, rt string) error
	GetSession(id int) (*sessions.Session, error)
	DeleteSessionFromDB(id int) error
}
