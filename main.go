package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/surodinsergey/golang-balance/db"
	"io/ioutil"
	"log"
	"net/http"
)

type Balance struct {
	ID   int   `json:"id"`
	Sum  int   `json:"sum"`
	User *User `json:"user_id"`
}

type User struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	DB := db.SetupDB()
	defer DB.Close()
	sqlStatement := `SELECT b.id, b.sum, u.id, u.lastname, u.firstname FROM balances as b LEFT JOIN users as u ON b.user_id=u.id WHERE user_id=$1;`
	row := DB.QueryRow(sqlStatement, params["id"])
	var balance Balance
	var user User
	err := row.Scan(&balance.ID, &balance.Sum, &user.ID, &user.Lastname, &user.Firstname)
	balance.User = &user
	switch err {
	case sql.ErrNoRows:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Нет баланса с id пользователя " + params["id"] + "!")
		return
	case nil:
		json.NewEncoder(w).Encode(balance)
		return
	default:
		w.WriteHeader(500)
		json.NewEncoder(w).Encode("Произошла серверная ошибка , попробуйте еще раз или обратитесь к администратору!")
		return
	}
}

func transferBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var (
		request map[string]map[string]int
	)
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Некоректные данные запроса!")
		return
	}
	json.Unmarshal(reqBody, &request)

	DB := db.SetupDB()
	defer DB.Close()
	sqlStatement := `SELECT b.id, b.sum FROM balances as b WHERE user_id=$1;`
	row := DB.QueryRow(sqlStatement, request["data"]["from"])

	var idFrom int
	var sumFrom int
	err = row.Scan(&idFrom, &sumFrom)

	switch err {
	case sql.ErrNoRows:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Нет баланса списания у пользователя с id " + string(request["data"]["from"]) + "!")
		return
	case nil:
		sqlStatement := `SELECT b.id, b.sum FROM balances as b WHERE user_id=$1;`
		row := DB.QueryRow(sqlStatement, request["data"]["to"])
		var idTo int
		var sumTo int
		err := row.Scan(&idTo, &sumTo)
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(404)
			json.NewEncoder(w).Encode("Нет баланса пополнения у пользователя с id " + string(request["data"]["to"]) + "!")
			return
		case nil:
			money := request["data"]["sum"]
			if checkBalance(sumFrom, money) {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode("На балансе списания у пользователя с id " + string(request["data"]["from"]) + " недостаточно средств!")
				return
			}
			sumFrom -= money
			sumTo += money
			_, err := DB.Exec("update balances set sum = $1 where id = $2", sumFrom, idFrom)
			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(err)
				return
			}
			_, err = DB.Exec("update balances set sum = $1 where id = $2", sumTo, idTo)
			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(err)
				return
			}
		default:
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(err)
			return
		}
	default:
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
		return
	}

	json.NewEncoder(w).Encode("Перевод успешно осуществлен!")
}

func updateBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var (
		request map[string]map[string]int
		balance Balance
		user    User
	)

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Некоректные данные запроса!")
		return
	}
	json.Unmarshal(reqBody, &request)

	DB := db.SetupDB()
	defer DB.Close()
	sqlStatement := `SELECT b.id, b.sum, u.id, u.lastname, u.firstname FROM balances as b LEFT JOIN users as u ON b.user_id=u.id WHERE user_id=$1;`
	row := DB.QueryRow(sqlStatement, params["id"])
	err = row.Scan(&balance.ID, &balance.Sum, &user.ID, &user.Lastname, &user.Firstname)
	balance.User = &user

	switch err {
	case sql.ErrNoRows:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Нет баланса с id пользователя " + params["id"] + "!")
		return
	case nil:
		if checkBalance(balance.Sum, request["data"]["sum"]) {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode("На балансе у пользователя с id " + string(params["id"]) + " недостаточно средств для списания!")
			return
		}
		balance.Sum += request["data"]["sum"]
		_, err = DB.Exec("update balances set sum = $1 where id = $2", balance.Sum, balance.ID)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(err)
			return
		}
	default:
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
		return
	}

	json.NewEncoder(w).Encode(balance)
}

func checkBalance(sum int, money int) bool {
	return sum-money < 0 || sum+money < 0
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	json.NewEncoder(w).Encode("Несуществующий метод API!")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/balance/{id}", getBalance).Methods("GET")
	r.HandleFunc("/balance/{id}", updateBalance).Methods("PUT")
	r.HandleFunc("/balance/transfer", transferBalance).Methods("POST")
	r.NotFoundHandler = http.HandlerFunc(notFound)
	log.Fatal(http.ListenAndServe(":8010", r))
}
