package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	handler *Handler
)

func connectToDB() (db *sql.DB, err error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s", os.Getenv("dbHostName"), os.Getenv("dbPort"), os.Getenv("dbUser"), os.Getenv("dbPass"), os.Getenv("dbName"))
	return sql.Open("postgres", psqlInfo)
	//fmt.Sprintf("root:root@tcp(localhost:3306)/arte?parseTime=true&charset=utf8mb4")
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

	conn, err := connectToDB()
	if err != nil {
		log.Fatalln(err)
	}

	sqlIni, err := ioutil.ReadFile("db.sql")
	if err != nil {
		log.Fatalln("error getting the db.sql file")
	}

	_, err = conn.Exec(string(sqlIni))
	if err != nil {
		log.Fatalln(err)
	}
	handler = NewHandler()
}
