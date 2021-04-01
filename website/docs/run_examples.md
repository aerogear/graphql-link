---
id: run_examples
title: Run Example GraphQL Servers
sidebar_label: Run Example GraphQL Servers
slug: /run_examples
---

To run the provided examples first go into examples folder.

```bash
$ cd examples
```

There are four examples present in the folder.

- Characters
- Shows
- Starwar Characters
- Starwar Starships

### Running Characters API

```bash
$ go run character.go

GraphQL endpoint running at http://127.0.0.1:8081/graphql
GraphQL UI running at http://127.0.0.1:8081
```

### Running Shows API

```bash
$ go run shows.go

GraphQL endpoint running at http://127.0.0.1:8082/graphql
GraphQL UI running at http://127.0.0.1:8082
```

### Running Starwar Characters API

```bash
$ go run starwars_characters.go

GraphQL endpoint running at http://127.0.0.1:8083/graphql
GraphQL UI running at http://127.0.0.1:8083
```

### Running Starwar Starships API

```bash
$ go run starwars_starships.go

GraphQL endpoint running at http://127.0.0.1:8084/graphql
GraphQL UI running at http://127.0.0.1:8084
```

Now Visit the end points to check the server is up

### Example

> When you visit <a href="http://127.0.0.1:8081/">http://127.0.0.1:8081/</a> add the following as an query to test it out.

##### Query

```graphql
# query to get id, name, friends, bestFriend and likes of all characters
query characters {
  characters {
    id
    name {
      first
      last
    }
    friends
    bestFriend
    likes
  }
}
```

##### Response

```json
// Sample JSON response
{
  "data": {
    "characters": [
      {
        "id": "1",
        "name": {
          "first": "Rukia",
          "last": "Kuchiki"
        },
        "friends": ["Ichigo", "Orihime"],
        "bestFriend": "Ichigo",
        "likes": 0
      },
      {
        "id": "2",
        "name": {
          "first": "Ichigo",
          "last": "Kurosaki"
        },
        "friends": [],
        "bestFriend": null,
        "likes": 0
      },
      {
        "id": "3",
        "name": {
          "first": "Orihime",
          "last": "Inoue"
        },
        "friends": [],
        "bestFriend": null,
        "likes": 0
      }
    ]
  }
}
```

Similarly you can test out other servers.
