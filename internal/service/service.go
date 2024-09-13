package service

import (
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
)

type Handlers interface {
	GetToken(id int) (string, error)
	SaveSession(id int, email, ip string, regTime time.Time, expireTime time.Duration, rt string) error
	GetSession(id int) (*database.MySession, error)
	DeleteSessionFromDB(id int) error
}

type Service struct {
	Handlers
}

func NewService(repo *database.Repo) *Service {
	return &Service{
		Handlers: repo,
	}
}
