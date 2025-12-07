package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	config := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode)
	db, err := sqlx.Open("postgres", config)
	if err != nil {
		fmt.Println("error to open postgres db")
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db); err != nil {
		fmt.Printf("Ошибка при применении миграций: %v\n", err)
		return nil, err
	}

	return db, nil
}

func (p *Postgres) AddAccount(user User) string {
	res, _ := p.DB.Query("SELECT username FROM users WHERE chat_id = $1", user.Chatid)
	tmp := ""
	res.Scan(&tmp)
	if tmp == "" {
		p.DB.Exec("INSERT INTO users (chat_id, username, requests, admin) VALUES ($1, $2, 5, 0)", user.Chatid, user.Username)
		return ""
	}

	return tmp
}
func (p *Postgres) CheckUser(chatId int64) bool {
	res := p.DB.QueryRow("SELECT chat_id FROM users WHERE chat_id = $1", chatId)
	check := ""
	res.Scan(&check)
	return check != ""
}
func (p *Postgres) CheckRequests(chatId int64) int {
	res := p.DB.QueryRow("SELECT requests FROM users WHERE chat_id = $1", chatId)
	requests := 0
	res.Scan(&requests)
	return requests
}

func (p *Postgres) MinusRequest(chatId int64) error {
	_, err := p.DB.Exec("UPDATE users SET requests = requests - 1 WHERE chat_id = $1", chatId)
	if err != nil {
		return err
	}
	return nil
}
