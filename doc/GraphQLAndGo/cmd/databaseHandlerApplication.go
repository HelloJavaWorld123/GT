package main

import (
	"database/sql"
	"log"
	"net/http"
)

func main() {
	db, err := sql.Open("postgresql", "....")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	http.Handle("/hello", &HelloHandler{db: db})
	http.ListenAndServe(":8080", nil)
}
