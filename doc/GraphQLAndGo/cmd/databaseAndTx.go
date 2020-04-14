package main

import (
	"GT/doc/GraphQLAndGo/myapp"
	"log"
)

func main() {
	//create one user
	db, err := myapp.Open("test")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	tx.CreatUser(&myapp.UserInfo{UserName: "test"})
	tx.Commit()

	//create batch users
	tx, err = db.Begin()
	userInfos := [10]myapp.UserInfo{}
	for _, u := range userInfos {
		tx.CreatUser(&u)
	}
	tx.Commit()
}
