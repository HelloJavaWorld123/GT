package myapp

import (
	"database/sql"
)

type DB struct {
	*sql.DB
}

type Tx struct {
	*sql.Tx
}

//open datasource
func Open(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgresql", dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DB{db}, err
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()

	if err != nil {
		return nil, err
	}
	return &Tx{tx}, err
}
