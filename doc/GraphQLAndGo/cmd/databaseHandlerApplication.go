package main

import (
	"GT/doc/GraphQLAndGo/structure"
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

	http.Handle("/hello", &structure.HelloHandler{Db: db})
	http.ListenAndServe(":8080", nil)
}
