# Structuring Applications In Go

#### Don't Use Global Variables
我读过的Go net/http 的列子大部分是用 http.HandleFunc,想下面这样:
(The Go net/http examples I read always show a function registered with http.HandleFunc like this:)
```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/hello",hello)
	err := http.ListenAndServe(":8080", nil)
	if err!= nil {
		fmt.Println(err)
	}
}


func hello(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("this is a Examples for Go Http ")
}
```
上面列子给了一种非常方便的方法使用net/http,但是是一种不好的习惯。通过使用函数处理器，