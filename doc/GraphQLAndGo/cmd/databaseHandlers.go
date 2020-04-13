package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

type HelloHandler struct {
	db *sql.DB
}

func (helloHandler *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var name string
	row := helloHandler.db.QueryRow("select myname from mytable")
	if err := row.Scan(&name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//Write it back to the client
	fmt.Fprintf(w, "hi %s \n", name)
}
