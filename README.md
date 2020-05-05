# graphql-gw

This is a GraphQL gateway that lets you easily compose or federate other
GraphQL endpoints.

<img src="https://raw.githubusercontent.com/chirino/graphql-gw/master/docs/images/graphql-gw-overview.jpg" alt="diagram of graphql-gw" width="488">

## Features

* Consolidate access to multiple upstream GraphQL endpoints via a single GraphQL gateway endpoint.
* The configuration uses GraphQL queries to define which upstream fields and types can be accessed.    
* Upstream types, that are accessible, are automatically merged into the gateway schema.
* Type conflict due to the same type name existing in multiple upstream endpoints can be avoided
  by renaming types in the gateway.
* Supports GraphQL Queries, Mutations, and Subscriptions

### Installing

`go get -u github.com/chirino/graphql-gw`

## Getting started

Run

```
graphql-gw new myproject
cd myproject
# edit myproject/graphql-gw.yaml to configure
# the gateway
graphql-gw serve
```

## Build from source

```bash
go build -o=graphql-gw main.go
```

### Usage

`$ graphql-gw --help`

```
A GraphQL composition gateway

Usage:
  graphql-gw [command]

Available Commands:
  help        Help about any command
  new         creates a graphql-gw project with default config
  serve       Runs the gateway service

Flags:
  -h, --help      help for graphql-gw
      --verbose   enables increased verbosity

Use "graphql-gw [command] --help" for more information about a command.
```

## Configuration Guide

### `listen:`

Set the `listen:` to the host and port you want the graphql endpoint to listen on. 

Example:
```yaml
listen: localhost:8080
```

### `endpoints:`

The endpoints section holds a map of all the upstream endpoint address that you will be
accessing with the gateway.  The example below defines two endpoints: `anilist` and `users`.
Keep in mind that the URL configured must be a graphql endpoint that is accessible from the 
gateway's network. 

```yaml
endpoints:
  anilist:
    url: https://graphql.anilist.co/
  users:
    url: https://users.acme.io/graphql
```

If there are duplicate types across the endpoints you can configure either type name
prefixes or suffixes on the endpoints so that conflicts can be avoided when imported
into the gateway.  Example:

```yaml
endpoints:
  anilist:
    url: https://graphql.anilist.co/
    prefix: Ani
    suffix: Type
  users:
    url: https://users.acme.io/graphql
```

### `types:`

Use the `types:` section of the configuration to define the fields that can be 
accessed by clients of the `graphql-gw`.  The root query and mutation type names 
are `Query` and `Mutation`.  Use those to configure the fields accessible in the 
root queries.  

The following example will add a field `myfield` to the `Query` type where the type
is the root query of the `anilist` endpoint. 

```yaml
types:
  - name: Query
    fields:
    - endpoint: anilist
      query: query {}
      name: myfield
```

`fields:` is a list of the following configuration elements:
 
| Field | Required| Description | 
|---|---| ---|
| `endpoint:` | yes | a reference to an endpoint defined in the `endpoints:` section.
| `query:` | yes | partial graphql query document to one node in the upstream endpoint graph.
| `name:` | no | field name to mount the resulting node on to.  not not specified, then all the field of the node are mounted on to the the parent type.|

### Importing all the fields of an upstream graphql endpoint.

If you want to import all the fields of an upstream endpoint object, simply don't specify 
the name for the field to mount the query on to. The following example will import all the
query fields onto the `Query` type and all the mutation fields on the `Mutation` type. 

```yaml
types:
  - name: Query
    fields:
    - endpoint: anilist
      query: query {}
  - name: Mutation
    fields:
    - endpoint: anilist
      query: mutation {}
```

### Importing a nested field of upstream graphql endpoint.

Use a full graphql query to access nested child graph elements of the upstream
graphql server.  Feel free to use argument variables or literals in the query. 
variables used in the query will be translated into arguments for the newly defined
field. 

```yaml
types:
  - name: Query
    fields:
    - endpoint: anilist
      query: |
        query ($search:String, $page:Int) {
          Page(page:$page, perPage:10) {
            characters(search:$search)
          }
        }
      name: pagedCharacterSearch
```

In the example above a `pagedCharacterSearch(search:String, page:Int)` field would 
be defined on the Query type and it would return the type returned by the `characters`
field of the anilist endpoint. 

## License

[BSD](./LICENSE)

## Development

- We love [pull requests](https://github.com/chirino/graphql-gw/pulls)
- [Open Issues](https://github.com/chirino/graphql-gw/issues)
- graphql-gw is written in [Go](https://golang.org/). It should work on any platform where go is supported.
- Built on this [GraphQL](https://github.com/chirino/graphql) framework
