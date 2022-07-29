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

func checkErr(err error) bool {
	return err != nil
}

// SetupDB DB set up
func SetupDB() (*sql.DB, error) {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DbHost, DbPort, DbUser, DbPassword, DbName)
	db, err := sql.Open("postgres", dbinfo)
	if checkErr(err) {
		return nil, err
	}
	return db, nil
}
