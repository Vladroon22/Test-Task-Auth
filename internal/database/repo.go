package database

import (
	"database/sql"
	"errors"
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
	ExpiresAt    time.Time
	RefreshToken string
}

func NewRepo(db *Storage) *Repo {
	return &Repo{
		sql: db,
	}
}
func (rp *Repo) GetToken(id int) (string, error) {
	var refreshToken string
	query := "SELECT refresh FROM sessions WHERE userID = $1"
	err := rp.sql.sql.QueryRow(query, id).Scan(&refreshToken)
	if err == sql.ErrNoRows {
		return "", err
	}
	if err != nil {
		log.Panicln(err)
		return "", err
	}
	return refreshToken, nil
}

func (rp *Repo) SaveSession(id int, email, ip string, expireAt time.Time, rt string) error {
	query := "INSERT INTO sessions (userID, email, userIP, expireAt, refresh) VALUES ($1, $2, $3, $4, $5)"
	if _, err := rp.sql.sql.Exec(query, id, email, ip, expireAt.Format(time.DateTime), rt); err != nil {
		log.Panicln(err)
		return err
	}
	return nil
}

func (rp *Repo) GetSession(id int) (*MySession, error) {
	if rp.sql.sql == nil {
		return nil, errors.New("database connection is nil")
	}

	session := &MySession{}
	query := "SELECT userID, email, userIP, expireAt, refresh FROM sessions WHERE userID = $1"
	err := rp.sql.sql.QueryRow(query, id).Scan(&session.UserID, &session.Email, &session.UserIP, &session.ExpiresAt, &session.RefreshToken)

	if err == sql.ErrNoRows {
		return nil, errors.New("no session found for userID")
	}

	if err != nil {
		return nil, err
	}
	return session, nil
}

func (rp *Repo) DeleteSessionFromDB(id int) error {
	query := "DELETE FROM sessions WHERE userID = $1"
	if _, err := rp.sql.sql.Exec(query, id); err != nil {
		return err
	}
	return nil
}
