# [Building an API With GraphQL And GO](https://medium.com/@bradford_hamilton/building-an-api-with-graphql-and-go-9350df5c9356)
>本博客中我们将使用**GO**、**GraphQL**、**PostgreSQL**创建一个API.我已在项目结构上迭代几个版本，这个是我最新欢的一个。在大部分的时间,我创建web APIs都是通过**Node.js**和**Ruby/Rails**.
>我发现第一次使用**Go**创建APIs时,需要费很大的劲儿。***Ben Johnson***的[Structuring Applications in Go ](/@benbjohnson/structuring-applications-in-go-3b04be4ff091)文章
>对我有很大的帮助,本博客中的部分代码就得益于***Ben Johnson***文章的指导,推荐阅读。


#### Setup
首先，先进行安装。在本篇博客中，我将在macOS中完成。
如果在你的macOS上还没有**Go**和**PostGreSQL**,[这片文章](/github.com/bradford-hamilton/go-graphql-api)详细讲解如何在**macOS**上配置**Go**和**PostgreSQL**.

创建一个新项目--**go-graphal-api**,整体项目结构如下：
```go
├── gql
│   ├── gql.go
│   ├── queries.go
│   ├── resolvers.go
│   └── types.go
├── main.go
├── postgres
│   └── postgres.go
└── server
    └── server.go
```

有一些额外依赖需要安装。我喜欢开发过程中能够热加载的[realize](https://github.com/oxequa/realize),go-chi的轻量级路由 [chi](https://github.com/go-chi/chi) 和 
管理request/response负载的 [render](https://github.com/go-chi/render) ,以及 [graphql-go/graphql](https://github.com/graphql-go/graphql) 作为我们的实现
```
go get github.com/oxequa/realize
go get github.com/go-chi/chi
go get github.com/go-chi/render
go get github.com/graphql-go/graphql
go get github.com/lib/pq
```

最后,我们创建一个数据库和一些需要测试的数据,运行**psql** 进行Postgres的命令行,创建一个数据库:
```postgresql
CREATE DATABASE go_graphql_db;
```

然后链接上该库：
```postgresql
\c go_graphql_db
```
链接上后,将一下sql语句粘贴到命令行:
```postgresql
CREATE TABLE users (
  id serial PRIMARY KEY,
  name VARCHAR (50) NOT NULL,
  age INT NOT NULL,
  profession VARCHAR (50) NOT NULL,
  friendly BOOLEAN NOT NULL
);

INSERT INTO users VALUES
  (1, 'kevin', 35, 'waiter', true),
  (2, 'angela', 21, 'concierge', true),
  (3, 'alex', 26, 'zoo keeper', false),
  (4, 'becky', 67, 'retired', false),
  (5, 'kevin', 15, 'in school', true),
  (6, 'frankie', 45, 'teller', true);
```

我们创建了一个基础的用户表并新增了6条新用户数据，对本博客来说已经足够.接下来开始构建我们的API!

#### API
在这篇博客中,所有的代码片段会都将包含一些注释,以帮助理解每一步.

从**main.go**开始:
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bradford-hamilton/go-graphql-api/gql"
	"github.com/bradford-hamilton/go-graphql-api/postgres"
	"github.com/bradford-hamilton/go-graphql-api/server"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
)

func main() {
	// Initialize our api and return a pointer to our router for http.ListenAndServe
	// and a pointer to our db to defer its closing when main() is finished
	router, db := initializeAPI()
	defer db.Close()

	// Listen on port 4000 and if there's an error log it and exit
	log.Fatal(http.ListenAndServe(":4000", router))
}

func initializeAPI() (*chi.Mux, *postgres.Db) {
	// Create a new router
	router := chi.NewRouter()

	// Create a new connection to our pg database
	db, err := postgres.New(
		postgres.ConnString("localhost", 5432, "bradford", "go_graphql_db"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create our root query for graphql
	rootQuery := gql.NewRoot(db)
	// Create a new graphql schema, passing in the the root query
	sc, err := graphql.NewSchema(
		graphql.SchemaConfig{Query: rootQuery.Query},
	)
	if err != nil {
		fmt.Println("Error creating schema: ", err)
	}

	// Create a server struct that holds a pointer to our database as well
	// as the address of our graphql schema
	s := server.Server{
		GqlSchema: &sc,
	}

	// Add some middleware to our router
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // set content-type headers as application/json
		middleware.Logger,          // log api request calls
		middleware.DefaultCompress, // compress results, mostly gzipping assets and json
		middleware.StripSlashes,    // match paths with a trailing slash, strip it, and continue routing through the mux
		middleware.Recoverer,       // recover from panics without crashing server
	)

	// Create the graphql route with a Server method to handle it
	router.Post("/graphql", s.GraphQL())

	return router, db
}
```

上面的**gql**、**postgres**、and**server**的路径应该是你本地的路径,以及**postgres.ConnString()** 链接PostgreSQL的用户名也应该是你自己的和我的不一样.
在 **initializeAPI()** 分为几大块主要的部分,接下来我们逐步构建每一块。

**chi.NewRouter()** 创建router并返回一个mux,接下来我们创建一个新的链接来链接我们的PostgreSQL数据库.
使用 **postgres.ConnString()** 创建一个***string*** 类型的链接配置,并封装到 **postgres.New()** 函数中.我们在自己包中的 **postgres.go** 文件中构建:
```go
package postgres

import (
	"database/sql"
	"fmt"

	// postgres driver
	_ "github.com/lib/pq"
)

// Db is our database struct used for interacting with the database
type Db struct {
	*sql.DB
}

// New makes a new database using the connection string and
// returns it, otherwise returns the error
func New(connString string) (*Db, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	// Check that our connection is good
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Db{db}, nil
}

// ConnString returns a connection string based on the parameters it's given
// This would normally also contain the password, however we're not using one
func ConnString(host string, port int, user string, dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbName,
	)
}

// User shape
type User struct {
	ID         int
	Name       string
	Age        int
	Profession string
	Friendly   bool
}

// GetUsersByName is called within our user query for graphql
func (d *Db) GetUsersByName(name string) []User {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := d.Prepare("SELECT * FROM users WHERE name=$1")
	if err != nil {
		fmt.Println("GetUserByName Preperation Err: ", err)
	}

	// Make query with our stmt, passing in name argument
	rows, err := stmt.Query(name)
	if err != nil {
		fmt.Println("GetUserByName Query Err: ", err)
	}

	// Create User struct for holding each row's data
	var r User
	// Create slice of Users for our response
	users := []User{}
	// Copy the columns from row into the values pointed at by r (User)
	for rows.Next() {
		err = rows.Scan(
			&r.ID,
			&r.Name,
			&r.Age,
			&r.Profession,
			&r.Friendly,
		)
		if err != nil {
			fmt.Println("Error scanning rows: ", err)
		}
		users = append(users, r)
	}

	return users
}
```
上面的思想是：创建数据库的链接并返回持有该数据库链接的**Db**对象.As you can tell the actual functionality/query is pretty useless, but the point is to get everything wired up as a good starting point.

将关注点重新回到 **main.go**文件,在40行处创建了一个root query 用于创建构建GraphQL的schema.我们在 **gql** 包下的 **queries.go** 中创建:
```go
package gql

import (
	"github.com/bradford-hamilton/go-graphql-api/postgres"
	"github.com/graphql-go/graphql"
)

// Root holds a pointer to a graphql object
type Root struct {
	Query *graphql.Object
}

// NewRoot returns base query type. This is where we add all the base queries
func NewRoot(db *postgres.Db) *Root {
	// Create a resolver holding our databse. Resolver can be found in resolvers.go
	resolver := Resolver{db: db}

	// Create a new Root that describes our base query set up. In this
	// example we have a user query that takes one argument called name
	root := Root{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					"users": &graphql.Field{
						// Slice of User type which can be found in types.go
						Type: graphql.NewList(User),
						Args: graphql.FieldConfigArgument{
							"name": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
						},
						Resolve: resolver.UserResolver,
					},
				},
			},
		),
	}
	return &root
}
```

接下来看一下 **resolvers.go**:
```go
package gql

import (
	"github.com/bradford-hamilton/go-graphql-api/postgres"
	"github.com/graphql-go/graphql"
)

// Resolver struct holds a connection to our database
type Resolver struct {
	db *postgres.Db
}

// UserResolver resolves our user query through a db call to GetUserByName
func (r *Resolver) UserResolver(p graphql.ResolveParams) (interface{}, error) {
	// Strip the name from arguments and assert that it's a string
	name, ok := p.Args["name"].(string)
	if ok {
		users := r.db.GetUsersByName(name)
		return users, nil
	}

	return nil, nil
}
```

打开 **types.go** :
```go
package gql

import "github.com/graphql-go/graphql"

// User describes a graphql object containing a User
var User = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"age": &graphql.Field{
				Type: graphql.Int,
			},
			"profession": &graphql.Field{
				Type: graphql.String,
			},
			"friendly": &graphql.Field{
				Type: graphql.Boolean,
			},
		},
	},
)
```
类似的,在这里添加我们不同的类型,每一个字段都指定了类型.在 **main.go**文件的42行使用root query创建了一个新的模式.


#### Almost there !
在**main.go**往下的51行处,创建一个新的server,server持有GraphQL的指针。下面是**server.go**的内容:
```go
package server

import (
	"encoding/json"
	"net/http"

	"github.com/bradford-hamilton/go-graphql-api/gql"
	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
)

// Server will hold connection to the db as well as handlers
type Server struct {
	GqlSchema *graphql.Schema
}

type reqBody struct {
	Query string `json:"query"`
}

// GraphQL returns an http.HandlerFunc for our /graphql endpoint
func (s *Server) GraphQL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check to ensure query was provided in the request body
		if r.Body == nil {
			http.Error(w, "Must provide graphql query in request body", 400)
			return
		}

		var rBody reqBody
		// Decode the request body into rBody
		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
		}

		// Execute graphql query
		result := gql.ExecuteQuery(rBody.Query, *s.GqlSchema)

		// render.JSON comes from the chi/render package and handles
		// marshalling to json, automatically escaping HTML and setting
		// the Content-Type as application/json.
		render.JSON(w, r, result)
	}
}
```
对于**GraphQL**方法，是处理GraphQL查询的唯一方法。注意:更新**gql**的路径是你本地的路径.
接下来看一下最后一个文件**gql.go**：
```go
package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

// ExecuteQuery runs our graphql queries
func ExecuteQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})

	// Error check
	if len(result.Errors) > 0 {
		fmt.Printf("Unexpected errors inside ExecuteQuery: %v", result.Errors)
	}

	return result
}
```
**ExecuteQuery()** 函数用来执行我们的GraphQL查询。我设想这里应该有一个类似于**ExecuteMutation()** 函数用来处理GraphQL的变化.

在**initializeAPI()**的最后增加一些中间工具以及增加处理**/graphql**请求的**GraphQL** server方法。在这里增加额外的RESTful路由，以及在server中增加处理的方法。

接下来在项目的根目录运行**realize init**,有两次提示信息并且两次都输入**n**

下面是在你项目的根目录下创建的 **.realize.yaml** 文件:
```
settings:
  legacy:
    force: false
    interval: 0s
schema:
- name: go-graphql-api
  path: .
  commands:
    run:
      status: true
  watcher:
    extensions:
    - go
    paths:
    - /
    ignored_paths:
    - .git
    - .realize
    - vendor
```
这段配置对于监控你项目里面的改变非常重要,如果检测到有改变,将自动重启server并重新运行 **main.go** 文件.

有一些暴露GraphQL API的非常好的工具,比如: **[graphiql](https://github.com/graphql/graphiql) **、**[insomnia](https://insomnia.rest/) **、**[graphql-playground](https://github.com/prisma/graphql-playground)**,
还可以发送一个application/json请求体的POST请求,比如:
```
{
 "query": "{users(name:\"kevin\"){id, name, age}}"
}
```

在[Postman](https://www.getpostman.com/) 里像下面这样:

在查询中只请求一个属性或者多个属性的组合.在GraphQL的正式版中,可以只请求我们希望通过网络发送的信息。


#### Great Success!
This's it!希望这篇博客对你在Go中编写GraphQL API有帮助。我尝试将功能分解到不同的包或文件中，使其更容易扩展,而且每一块也很容易测试。