package structure

import (
	"bytes"
	"database/sql"
	"net/http/httptest"
	"testing"
)

func TestHelloHandler_ServeHTTP(t *testing.T) {
	db, _ := sql.Open("postgresql", ".....")
	defer db.Close()
	handler := HelloHandler{Db: db}

	recorder := httptest.NewRecorder()
	recorder.Body = bytes.NewBuffer(make([]byte, 10))
	handler.ServeHTTP(recorder, nil)
	if recorder.Body.String() != "hi bob!\n" {
		t.Error("unexpected response: %s", recorder.Body.String())
	}
}
