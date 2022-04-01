package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	handler *Handler
)

func connectToDB() (db *sql.DB, err error) {
	return sql.Open("mysql", "root:root@tcp(localhost:3306)/arte?parseTime=true&charset=utf8mb4")
}

// func removeFromSlice[T comparable](slice []T, i T) []T {
// 	var toReturn []T
// 	for _, v := range slice {
// 		if v != i {
// 			toReturn = append(toReturn, v)
// 		}
// 	}
// 	return toReturn
// }

func removeFromSlice(slice []int, i int) []int {
	var toReturn []int
	for _, v := range slice {
		if v != i {
			toReturn = append(toReturn, v)
		}
	}
	return toReturn
}

func init() {
	//load the enviroment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	handler = NewHandler()
}
