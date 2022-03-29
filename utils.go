package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

func connectToDB() (db *sql.DB, err error) {
	return sql.Open("mysql", "root:root@tcp(localhost:3306)/arte?parseTime=true&charset=utf8mb4")
}

func init() {
	//load the enviroment variables
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}
