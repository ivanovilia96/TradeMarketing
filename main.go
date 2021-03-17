package main

import (
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	defer ConnectedDataBase.Close()
	println("You can work on http://localhost:8080/ ")
	r := mux.NewRouter()
	r.HandleFunc("/statistics-from={fromDate},to={toDate},sortField={sortField}", GetStatistics).Methods(http.MethodGet)
	r.HandleFunc("/statistics", PutStatistics).Methods(http.MethodPut)
	r.HandleFunc("/statistics", DeleteStatistics).Methods(http.MethodDelete)
	log.Fatal(http.ListenAndServe(":8080", r))
}
