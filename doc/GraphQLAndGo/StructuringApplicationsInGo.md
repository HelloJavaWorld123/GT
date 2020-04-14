# [Structuring Applications In Go](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091)

#### Don't Use Global Variables
我读过关于Go **net/http** 的列子大部分是用**http.HandleFunc**,像下面这样:
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
从根目录下移除**main.go**文件，有助于从库的角度构建你的应用。应用程序的二进制文件只是你项目库的一个简单客户端。
(Moving the **main.go** file out of your root allows you build your application from the perspective of a library.Your application binary is simply a client of your application's library)
我发现可以帮助我很清晰的抽象出那些代码是库的核心逻辑以及那些代码是运行这个项目
(I find this helps me make a cleaner abstraction of what code is for my core logic(the library) and what code is for running my application(the application binary))

项目的二进制文件只是用户与你的逻辑交互的入口。有时你想要有多种交互方式因此你创建了多个二进制文件。
比如：你有一个“adder”包可以让用户所有的数字进行累加，你可能想要发布一个命令行版本和web版本。通过下面这种方法就很容易做到：
```go
adder/
  adder.go
  cmd/
    adder/
      main.go
    adder-server/
      main.go
```
用户可以使用 **go get**命令安装你的“adder”应用：
```
$ go get github.com/benbjohnson/adder/...
```

现在,用户将安装了 "adder" 和 “adder-server”

#### Wrap types for application-specific context
我发现一个特别有用的技巧，包装一些泛型类型作为应用程序的上下文。一个很好的列子就是包装DB和Tx的类型。
这些类型可以在**database/sql**包或者其他数据库中比如：[Bolt](https://github.com/boltdb/bolt)

开始像下面这样在我们项目中包装DB和Tx类型:
```go
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
```

然后创建数据库和事务初始化函数：
```go
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
	db,err := sql.Open("postgresql",dataSourceName)
	if err != nil {
		return nil, err
	}
	
	return &DB{db}, err
}

func (db *DB) Begin()(*Tx,error)  {
	tx,err := db.DB.Begin()

	if err != nil {
		return nil, err
	}
	return &Tx{tx},err
}
```
使用事务为我们的项目创建具体的函数，比如：项目中有创建用户并且创建用户之前需要增加验证,那么创建一个**Tx.CreateUser**函数,如下：
```go
package myapp

import (
	"database/sql"
	"errors"
)

func (tx *Tx) CreatUser(info *UserInfo) (sql.Result,error){
	if info == nil {
		return nil,errors.New("user must not bee null")
	}

	if info.UserName == "" {
		return nil,errors.New("userName must not bee null")
	}

	result, err := tx.Exec("insert into user_info values (....)", "")
	if err != nil {
		return nil,err
	}
	return result, err
}
```
比如：一个用户创建之前需要对另外一个系统进行验证或者另外一张表需要更新，那么这个函数将会变得更加复杂.但是对于应用程序的调用者，它将会被封装在一个函数中。

###### Transactional composition
将这些函数添加到Tx中的另一个好处是，可以在一个事务中完成多个操作。
比如：需要添加一个用户？ 调用一次**Tx.CreateUser()**
```go
tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	tx.CreatUser(&myapp.UserInfo{UserName: "test"})
	tx.Commit()
```
需要添加多个用户？也可是使用这个函数，而不需要创建**Tx.CreateUsers()**:
```go
tx, err = db.Begin()
for _,u := range userInfos {
    tx.CreatUser(&u)
}
tx.Commit()
```

对底层数据进行抽象还可以简化数据库之间的交换或者多个数据的操作。都被封装到项目中**DB & Tx**的类型相关的函数中。

#### Don’t go crazy with subpackages
大部分的语言支持用户按照自己喜欢的方式组织包的结构。我曾经使用Java库开发，java库中耦合的类可以塞进其他的包中，并且这些包可以相互包含。it was a mess(一团糟)

Go 对包的要求只有一个：不能循环依赖；刚开始我对循环依赖很陌生。
刚开始组织包时每一个文件只有一个类型，一旦在一个包下有很多文件时，我就创建一个子文件夹。
然而，当我不能在A包中包含B包，B包中包含C包,C包中包含A包时,这些子文件夹变得很难管理。这就是循环依赖。
我意识到，除了有太多文件之外，没有很好的利用把包分开。

最近我发现另外一个方向：只用一个根目录。
通常我得项目类型都是非常相关的，所以从可用性和API角度来看，它更合适。这些类型还可以利用在他们之间未导出的API来简化API。

我发现一些方法帮助我趋向于创建更大的包结构(large packages)：

- 将关联的类型和代码组织在一个文件中。如果类型和函数组织的很好，就会发现这个文件将趋向于200到500行。
这听起来很大，但却很容易关联查找(navigate)。我的文件的上限是1000行。
- 将重要的类型放在文件的顶部，不重要的类型放在文件的底部。
- 一旦你的项目源码超过10000行，你需要认真的评估是否需要分割成更小的工程。

[Bolt](https://github.com/boltdb/bolt) 是一个很好的列子。每个文件都是与Bolt相关的类型的分组。
```
bucket.go
cursor.go
db.go
freelist.go
node.go
page.go
tx.go
```

#### Conclusion
组织代码是软件编程中很重要的一部分，但是却很少得到应有的关注。谨慎使用全局变量、将**main.go**移动到指定的文件夹、Wrap types for application-specific context、
以及减少子包的数量，这些只是使Go项目在易用性和扩展性上的一部分技巧。

如果使用编写Go项目同样的方法编写java、Ruby或者Node.js会很难进行。