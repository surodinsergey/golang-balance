package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Balance struct {
	ID   int   `json:"id"`
	Sum  int   `json:"sum"`
	User *User `json:"user"`
}

type User struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type RequestData struct {
	Data string `json:"data"`
}

var balances []Balance

func getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, item := range balances {
		if strconv.Itoa(item.User.ID) == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Balance{}) //TODO: Заменить на ошибку
}

func transferBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var (
		request map[string]map[string]int
	)
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Enter data!!")
	}
	json.Unmarshal(reqBody, &request)
	var (
		from,
		to int
	)
	for index, item := range balances {
		if item.User.ID == request["data"]["from"] {
			from = index
		}
		if item.User.ID == request["data"]["to"] {
			to = index
		}
	}

	balances[from].Sum -= request["data"]["sum"]
	balances[to].Sum += request["data"]["sum"]

	json.NewEncoder(w).Encode(balances)
}

func updateBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var (
		result  Balance
		request map[string]map[string]int
	)
	for index, item := range balances {
		if strconv.Itoa(item.User.ID) == params["id"] {
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, "Enter data!!")
			}
			json.Unmarshal(reqBody, &request)
			balances[index].Sum += request["data"]["sum"]
			result = balances[index]
			json.NewEncoder(w).Encode(result)
			return
		}
	}
	json.NewEncoder(w).Encode(&Balance{}) //TODO: Заменить на ошибку
}

func main() {
	r := mux.NewRouter()
	//TODO: интегрировать БД , обработать ошибки
	balances = append(balances, Balance{ID: 1, Sum: 100, User: &User{ID: 1, Firstname: "Василий", Lastname: "Пуговкин"}})
	balances = append(balances, Balance{ID: 2, Sum: 500, User: &User{ID: 2, Firstname: "Андрей", Lastname: "Серебров"}})
	r.HandleFunc("/balance/{id}", getBalance).Methods("GET")
	r.HandleFunc("/balance/{id}", updateBalance).Methods("PUT")
	r.HandleFunc("/balance/transfer", transferBalance).Methods("POST")
	log.Fatal(http.ListenAndServe(":8010", r))
}
