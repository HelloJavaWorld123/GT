# [Building an API With GraphQL And GO](https://medium.com/@bradford_hamilton/building-an-api-with-graphql-and-go-9350df5c9356)
>在这边博客中我们将使用**GO**、**GraphQL**、**PostgreSQL**创建一个API.我已经在项目结构上
>迭代了几个版本了，这个是我最新欢的一个。在大部分的时间了,我创建web APIs都是通过**Node.js**和**Ruby/Rails**.
>我发现第一次使用**Go**创建APIs时,需要费很大的劲儿。***Ben Johnson***的[Structuring Applications in Go ](/@benbjohnson/structuring-applications-in-go-3b04be4ff091)文章
>对我有很大的帮助,博客中的部分代码就得益于文章的直到,推荐阅读。


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

有一些额外依赖需要安装。我喜欢开发过程中能够热加载的[realize](https://github.com/oxequa/realize),
