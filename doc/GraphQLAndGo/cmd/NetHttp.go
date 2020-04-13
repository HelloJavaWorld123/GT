package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", hello)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func hello(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("this is a Examples for Go Http ")
}
