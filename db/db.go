package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	DbUser     = "go_balance"
	DbPassword = "secret"
	DbName     = "go_balance"
	DbHost     = "database"
	DbPort     = 5432
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// SetupDB DB set up
func SetupDB() *sql.DB {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DbHost, DbPort, DbUser, DbPassword, DbName)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	return db
}
