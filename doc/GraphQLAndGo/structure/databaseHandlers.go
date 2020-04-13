package structure

import (
	"database/sql"
	"fmt"
	"net/http"
)

type HelloHandler struct {
	Db *sql.DB
}

func (helloHandler *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var name string
	row := helloHandler.Db.QueryRow("select myname from mytable")
	if err := row.Scan(&name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//Write it back to the client
	fmt.Fprintf(w, "hi %s \n", name)
}
