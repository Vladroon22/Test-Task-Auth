package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	c "github.com/Vladroon22/Test-Task-BackDev/config"
)

type Storage struct {
	conf *c.Config
	sql  *sql.DB
}

func NewDB(conf *c.Config) *Storage {
	return &Storage{
		conf: conf,
	}
}

func (db *Storage) Connect() error {
	if err := db.configurate(*db.conf); err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}

func (db *Storage) configurate(cnf c.Config) error {
	source := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?%s", cnf.Username, cnf.Password, cnf.Host, cnf.Port, cnf.DBname, cnf.SSLmode)
	log.Println(source)
	sqlConn, err := sql.Open("postgres", source)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	if err := RetryPing(sqlConn); err != nil {
		log.Fatalln(err)
		return err
	}
	db.sql = sqlConn

	return nil
}

func RetryPing(sqlConn *sql.DB) error {
	var err error
	for i := 0; i < 5; i++ {
		if err = sqlConn.Ping(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func (db *Storage) CloseDB() error {
	return db.sql.Close()
}
