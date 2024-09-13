package database

import (
	"database/sql"
	"log"
	"time"
)

type Repo struct {
	sql *Storage
}

type MySession struct {
	UserID       int
	Email        string
	UserIP       string
	RegTime      time.Time
	RefreshToken string
	ExpireTime   time.Duration
}

func NewRepo(db *Storage) *Repo {
	return &Repo{
		sql: db,
	}
}
func (rp *Repo) GetToken(id int) (string, error) {
	var refreshToken string
	query := "SELECT refresh FROM sessions WHERE id = $1"
	err := rp.sql.sql.QueryRow(query, id).Scan(&refreshToken)
	if err == sql.ErrNoRows || err != nil {
		log.Panicln(err)
		return "", err
	}
	return refreshToken, nil
}

func (rp *Repo) SaveSession(id int, email, ip string, regTime time.Time, expireTime time.Duration, rt string) error {
	query := "INSERT INTO sessions (userID, email, userIP, regTime, expireTime, refresh) VALUES ($1, $2, $3, $4, $5, $6)"
	if _, err := rp.sql.sql.Exec(query, id, email, ip, regTime.Format(time.DateTime), expireTime, rt); err != nil {
		log.Panicln(err)
		return err
	}
	return nil
}

func (rp *Repo) GetSession(id int) (*MySession, error) {
	session := &MySession{}
	query := "SELECT FROM sessions WHERE id = $1"
	err := rp.sql.sql.QueryRow(query, id).Scan(&session.UserID, &session.Email, &session.UserIP, &session.RegTime, &session.RefreshToken, &session.ExpireTime)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (rp *Repo) DeleteSessionFromDB(id int) error {
	query := "DELETE FROM sessions WHERE id = $1"
	if _, err := rp.sql.sql.Exec(query, id); err != nil {
		return err
	}
	return nil
}
