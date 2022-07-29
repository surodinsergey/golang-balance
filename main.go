package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/surodinsergey/golang-balance/db"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Tranzaction struct {
	ID    int       `json:"id"`
	Sum   int       `json:"sum"`
	Type  string    `json:"type"`
	Date  time.Time `json:"date"`
	*User `json:"user_id"`
}

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
	DB, errDb := db.SetupDB()
	if errDb != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Errorf("Ошибка соединения с БД : %q", errDb.Error()).Error())
		return
	}
	defer DB.Close()
	var user User
	balance := &Balance{}

	sqlStatement := `SELECT u.id, u.lastname, u.firstname FROM users as u WHERE id=$1;`
	row := DB.QueryRow(sqlStatement, params["id"])
	err := row.Scan(&user.ID, &user.Lastname, &user.Firstname)
	switch err {
	case sql.ErrNoRows:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Нет пользователя c id " + params["id"] + "!")
		return
	case nil:
		balance.User = &user
		sqlStatement := `SELECT t.id, t.sum, t.type, t.date FROM tranzactions as t WHERE t.user_id=` + params["id"] + `;`
		rows, err := DB.Query(sqlStatement)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(fmt.Errorf("Ошибка сервера : %q", err.Error()).Error())
			return
		}

		for rows.Next() {
			var tranz Tranzaction

			err = rows.Scan(&tranz.ID, &tranz.Sum, &tranz.Type, &tranz.Date)

			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(fmt.Errorf("Ошибка сервера : %q", err.Error()).Error())
				return
			}
			balance.Sum += tranz.Sum
		}
		balance.ID = user.ID
	default:
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
		return
	}

	json.NewEncoder(w).Encode(balance)
	return
}

func transferBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type TransferUsersBalanceRequest struct {
		Data struct {
			From int
			To   int
			Sum  int
		}
	}
	var (
		request     TransferUsersBalanceRequest
		userFrom    User
		userTo      User
		balanceFrom Balance
		balanceTo   Balance
	)
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Некоректные данные запроса!")
		return
	}
	json.Unmarshal(reqBody, &request)

	DB, errDb := db.SetupDB()
	if errDb != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(errDb)
		return
	}
	defer DB.Close()

	sqlStatement := `SELECT u.id, u.lastname, u.firstname FROM users as u WHERE id=$1;`
	row := DB.QueryRow(sqlStatement, request.Data.From)
	err = row.Scan(&userFrom.ID, &userFrom.Lastname, &userFrom.Firstname)
	switch err {
	case sql.ErrNoRows:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Нет пользователя c id " + strconv.Itoa(request.Data.From) + "!")
		return
	case nil:
		balanceFrom.User = &userFrom
		sqlStatement := `SELECT u.id, u.lastname, u.firstname FROM users as u WHERE id=$1;`
		row := DB.QueryRow(sqlStatement, request.Data.To)
		err = row.Scan(&userTo.ID, &userTo.Lastname, &userTo.Firstname)
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(404)
			json.NewEncoder(w).Encode("Нет пользователя c id " + strconv.Itoa(request.Data.To) + "!")
			return
		case nil:
			balanceTo.User = &userTo
			sqlStatement := `SELECT t.id, t.sum, t.type, t.date FROM tranzactions as t WHERE t.user_id=` + strconv.Itoa(request.Data.From) + `;`
			rows, err := DB.Query(sqlStatement)
			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(fmt.Errorf("Ошибка сервера : %q", err.Error()).Error())
				return
			}
			for rows.Next() {
				var tranz Tranzaction

				err = rows.Scan(&tranz.ID, &tranz.Sum, &tranz.Type, &tranz.Date)

				if err != nil {
					w.WriteHeader(500)
					json.NewEncoder(w).Encode(fmt.Errorf("Ошибка сервера : %q", err.Error()).Error())
					return
				}
				balanceFrom.Sum += tranz.Sum
			}
			money := -request.Data.Sum
			if checkBalance(balanceFrom.Sum, money) {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode("На балансе списания у пользователя с id " + strconv.Itoa(request.Data.From) + " недостаточно средств!")
				return
			}

			var (
				dateT        time.Time
				lastInsertID int
			)

			dateT = time.Now()

			err = DB.QueryRow("INSERT INTO tranzactions(sum, type, date, user_id) VALUES($1, $2, $3, $4) returning id;", -request.Data.Sum, "transfer", dateT, request.Data.From).Scan(&lastInsertID)

			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
				return
			}
			balanceFrom.ID = userFrom.ID

			err = DB.QueryRow("INSERT INTO tranzactions(sum, type, date, user_id) VALUES($1, $2, $3, $4) returning id;", request.Data.Sum, "transfer", dateT, request.Data.To).Scan(&lastInsertID)

			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
				return
			}
			balanceTo.ID = userTo.ID

		default:
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
			return
		}
	default:
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
		return
	}

	json.NewEncoder(w).Encode("Перевод успешно осуществлен!")
}

func updateBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	type PutUserBalanceRequest struct{ Data struct{ Sum int } }
	var (
		request PutUserBalanceRequest
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

	DB, errDb := db.SetupDB()
	if errDb != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(errDb)
		return
	}
	defer DB.Close()
	sqlStatement := `SELECT u.id, u.lastname, u.firstname FROM users as u WHERE id=$1;`
	row := DB.QueryRow(sqlStatement, params["id"])
	err = row.Scan(&user.ID, &user.Lastname, &user.Firstname)
	balance.User = &user

	switch err {
	case sql.ErrNoRows:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode("Нет пользователя c id " + params["id"] + "!")
		return
	case nil:
		sqlStatement := `SELECT t.id, t.sum, t.type, t.date FROM tranzactions as t LEFT JOIN users as u ON t.user_id=u.id WHERE t.user_id=` + params["id"] + `;`
		rows, err := DB.Query(sqlStatement)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(fmt.Errorf("Ошибка сервера : %q", err.Error()).Error())
			return
		}

		for rows.Next() {
			var tranz Tranzaction
			err = rows.Scan(&tranz.ID, &tranz.Sum, &tranz.Type, &tranz.Date)

			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(fmt.Errorf("Ошибка сервера : %q", err.Error()).Error())
				return
			}
			balance.Sum += tranz.Sum
		}
		if checkBalance(balance.Sum, request.Data.Sum) {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode("На балансе у пользователя с id " + string(params["id"]) + " недостаточно средств для списания!")
			return
		}
		balance.Sum += request.Data.Sum

		var (
			typeT        string
			dateT        time.Time
			lastInsertID int
		)

		if request.Data.Sum < 0 {
			typeT = "buy"
		} else {
			typeT = "pay"
		}

		dateT = time.Now()

		err = DB.QueryRow("INSERT INTO tranzactions(sum, type, date, user_id) VALUES($1, $2, $3, $4) returning id;", request.Data.Sum, typeT, dateT, params["id"]).Scan(&lastInsertID)

		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
			return
		}
		balance.ID = user.ID
	default:
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Errorf("Произошла серверная ошибка : : %q", err.Error()).Error())
		return
	}

	json.NewEncoder(w).Encode(balance)
}

func checkBalance(sum int, money int) bool {
	return sum+money < 0
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
