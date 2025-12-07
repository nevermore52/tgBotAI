package database

import "github.com/jmoiron/sqlx"

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}
type Postgres struct {
	DB *sqlx.DB
}

type User struct {
	Chatid 		int64
	Username	string
	Requests	int
	Admin		int
}
