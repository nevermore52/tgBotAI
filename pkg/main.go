package main

import (
	"fmt"
	"os"
	"tgbot/internal/database"
	tgbot "tgbot/internal/telegramApi"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	db, err := database.NewPostgresDB(database.Config{
	Host:		os.Getenv("POSTGRES_HOST"),
	Port:	 	os.Getenv("POSTGRES_PORT"),
	Username: 	os.Getenv("POSTGRES_USER"),
	Password:	os.Getenv("POSTGRES_PASSWORD"),
	DBName:		os.Getenv("POSTGRES_DBNAME"),
	SSLMode: 	"disable",
	})
		if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	
	pg := database.Postgres{DB: db}
	Tg := tgbot.NewTgBot(pg)
	
	Tg.StartBot() 
} 