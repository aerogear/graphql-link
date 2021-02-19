---
id: run_examples
title: Run Example GraphQL Servers
sidebar_label: Run Example GraphQL Servers
slug: /run_examples
---

### Go into the examples directory

```bash
$ cd examples
```

### Run example.go file

```bash
$ go run example.go

2021/02/20 00:53:55 ===== Characters =====
GraphQL endpoint running at http://127.0.0.1:8081/graphql
GraphQL UI running at http://127.0.0.1:8081
2021/02/20 00:53:55 ===== Shows =====
GraphQL endpoint running at http://127.0.0.1:8082/graphql
GraphQL UI running at http://127.0.0.1:8082
2021/02/20 00:53:55 ===== Starwars Characters =====
GraphQL endpoint running at http://127.0.0.1:8083/graphql
GraphQL UI running at http://127.0.0.1:8083
2021/02/20 00:53:55 ===== Starwars StarShip =====
GraphQL endpoint running at http://127.0.0.1:8084/graphql
GraphQL UI running at http://127.0.0.1:8084
```

Now Visit the end points to check the server is up

### Example

- When you visit <a href="http://localhost:8081">http://localhost:8081/</a> add the following as an query to test it out.

```graphql
query characters {
  characters {
    id
  }
}
```

```json
{
  "data": {
    "characters": [
      {
        "id": "1"
      },
      {
        "id": "2"
      },
      {
        "id": "3"
      },
      {
        "id": "3"
      },
      {
        "id": "3"
      },
      {
        "id": "3"
      }
    ]
  }
}
```

- Similarly other servers will also be running
