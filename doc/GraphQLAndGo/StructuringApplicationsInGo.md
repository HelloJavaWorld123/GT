# [Structuring Applications In Go](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091)

#### Don't Use Global Variables
我读过的Go **net/http** 的列子大部分是用 http.HandleFunc,想下面这样:
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
上面列子给了一种非常方便的方法使用 **net/http**,但是是一种不好的习惯。通过使用函数处理器，访问应用程序的唯一方法是使用全局变量。
(This examples gives an easy way to get into using net/http but it teaches a bad habit. By using a function handler,the only to access application state is to use a global variable)

为此,你可能会定义一个全局的数据库连接或者全局的配置变量,但是当写单元测试使用这些变量时将会是噩梦
(Because of this,you may decide to add a global database connection or a global configuration variable but these globals are a nightmare to use when write a unit test)

一个更好的方法是为处理器创建指定的类型，而且可以包含必要的变量:
```go
package main

import (
	"database/sql"
	"fmt"
	"net/http"
)
type HelloHandler struct {
	db *sql.DB
}

func (helloHandler *HelloHandler)  ServeHTTP(w http.ResponseWriter,r *http.Request)  {
	var name string
	row := helloHandler.db.QueryRow("select myname from mytable")
	if err :=row.Scan(&name);err != nil{
		http.Error(w,err.Error(),500)
		return
	}
	//Write it back to the client
	fmt.Fprintf(w,"hi %s \n",name)
}
```
现在我们在不使用全局变量条件下初始化数据库以及注册处理器
(now we can initialize our database and register our handler without the use of global variables)
```go
package main
import (
	"database/sql"
	"log"
	"net/http"
)
func main() {
	db,err :=sql.Open("postgresql","....")
	if err != nil{
		log.Fatal(err.Error())
	}

	http.Handle("/hello",&HelloHandler{db:db})
	http.ListenAndServe(":8080",nil)
}
```

这种方法还有一个益处，我们处理器的单元测试是独立的，甚至不需要一个Http服务器:
(This approach also has the benefit that unit testing our handler is self-contained and doesn't even require HTTP server:)
```go
package main

import (
	"bytes"
	"database/sql"
	"net/http/httptest"
	"testing"
)

func TestHelloHandler_ServeHTTP(t *testing.T) {
	// open our connection and setup our handler
	db,_ := sql.Open("postgresql",".....")
	defer db.Close()
	handler := HelloHandler{db: db}

    //executor our handler with a simple buffer
	recorder := httptest.NewRecorder()
	recorder.Body = bytes.NewBuffer(make([]byte,10))

	handler.ServeHTTP(recorder,nil)
	if recorder.Body.String() != "hi bob!\n" {
		t.Error("unexpected response: %s", recorder.Body.String())
	}
}
```

#### Separate your binary from your application
过去我把**main.go**放置在项目的更目录,当其他人运行**go get**时,我的项目会自动的安装。
然而，在同一个包下结合**main.go**文件和应用程序逻辑，导致两个结果：
- 使我的应用程序不能作为一个库使用
(it makes my application unusable as a library)
- 在一个应用程序内只能有一个**main.go**文件
(I can only one application binary)

我找到解决这个问题最好的办法是在项目里使用**cmd**文件夹,它的每一个子文件夹是一个工程的**main.go**.
(The best way I've found to fix this is to simply use a "cmd" directory in my project
where each of its subdirectories is an application binary) 
我最早发现使用这个方法是在***Fitzpatrick’s *** [Camlistore](http://camlistore.org/) 这个项目里,使用了很多项目二进制文件
(I originally found this approach used in **Fitzpatrick’s** [Camlistore](http://camlistore.org/) project 
where he uses several application binaries:)
```
camlistore/
  cmd/
    camget/
      main.go
    cammount/
      main.go
    camput/
      main.go
    camtool/
      main.go
```

当[Camlistore](http://camlistore.org/) 被安装的时候，可以有4个被分开的项目二进制文件可以被编译：
camget, cammount, camput, & camtool.
(Here we have 4 separate application binaries that can be build when [Camlistore](http://camlistore.org/) is installed:
camget, cammount, camput, & camtool.)

###### Library driven development
