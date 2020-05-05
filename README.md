# graphql-gw

This is a GraphQL gateway that lets you easily compose or federate other
GraphQL upstream servers.

<img src="https://raw.githubusercontent.com/chirino/graphql-gw/master/docs/images/graphql-gw-overview.jpg" alt="diagram of graphql-gw" width="488">

## Features

* Consolidate access to multiple upstream GraphQL servers via a single GraphQL gateway server.
* Introspection of the upstream server to discover their GraphQL schemas.
* The configuration uses GraphQL queries to define which upstream fields and types can be accessed.    
* Upstream types, that are accessible, are automatically merged into the gateway schema.
* Type conflict due to the same type name existing in multiple upstream servers can be avoided
  by renaming types in the gateway.
* Supports GraphQL Queries, Mutations, and Subscriptions
* Production mode settings to avoid the gateway's schema from dynamically changing due to 
  changes in the upstream schemas.  

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

### Development and Production Mode

The `graphql-gw serve` command will run the gateway in development mode.  Development mode
will cause the gateway to download upstream schemas on start up.  The schema files will
be stored in the `upstreams` directory (located in the same directory as the gateway 
configuration file).  If any of the schemas cannot be downloaded the gateway will fail to
startup.

You can use `graphql-gw serve --production` to enabled production mode.  In this mode, the 
schema for the upstream severs will be loaded from the `upstreams` directory that they were
stored when you used development mode.  This ensures that your gateway will have a consistent 
schema presented, and that it's start up will not be impacted by the availability of the upstream
servers.

## Configuration Guide

### `listen:`

Set the `listen:` to the host and port you want the graphql server to listen on. 

Example:
```yaml
listen: localhost:8080
```

### `upstreams:`

The upstreams section holds a map of all the upstream severs that you will be
accessing with the gateway.  The example below defines two upstream servers: `anilist` and `users`.
Keep in mind that the URL configured must be a graphql server that is accessible from the 
gateway's network. 

```yaml
upstreams:
  anilist:
    url: https://graphql.anilist.co/
  users:
    url: https://users.acme.io/graphql
```

If there are duplicate types across the upstream severs you can configure either type name
prefixes or suffixes on the upstream severs so that conflicts can be avoided when imported
into the gateway.  Example:

```yaml
upstreams:
  anilist:
    url: https://graphql.anilist.co/
    prefix: Ani
    suffix: Type
  users:
    url: https://users.acme.io/graphql
```

### `types:`

Use the `types:` section of the configuration to define the fields that can be 
accessed by clients of the `graphql-gw`.  The root query, mutation and subscription 
type names are `Query`, `Mutation`, `Subscription`.  Use those to configure the fields 
accessible from the root queries.  

The following example will add a field `myfield` to the `Query` type where the type
is the root query of the `anilist` upstream server. 

```yaml
types:
  - name: Query
    actions:
    - type: mount
      field: myfield
      upstream: anilist
      query: query {}
```

### `actions:`

`actions:` is a list configuration actions to take against on the named type.  The actions are processed in 
order.  You can select from the following actions types:

| type |Description | 
|---|---|
| [`mount`](#action-type-mount) | mounts an upstream field onto a gateway schema type using a graphql query
| [`rename`](#action-type-rename) | renames either a type or field in the gateway schema.

### Action `type: mount`

The `mount` action can be used to mount an upstream field onto a gateway schema type using a graphql query
 
| Field | Required| Description | 
|---|---| ---|
| `upstream:` | yes | a reference to an upstream server defined in the `upstreams:` section.
| `query:` | yes | partial graphql query document to one node in the upstream server graph.
| `field:` | no | field name to mount the resulting node on to.  not not specified, then all the field of the node are mounted on to the the parent type.|

### Action `type: rename`

The `rename` action can be used to rename either a type or field in the gateway schema. 
 
| Field | Required| Description | 
|---|---| ---|
| `field:` | no | if not set, you will be renaming the type, if set, you will be renaming a field of the type.
| `to:` | yes | the new name  |

## Common Use Cases

### Importing all the fields of an upstream graphql upstream server.

If you want to import all the fields of an upstream server type, simply don't specify 
the name for the field to mount the query on to. The following example will import all the
query fields onto the `Query` type and all the mutation fields on the `Mutation` type. 

```yaml
types:
  - name: Query
    actions:
    - type: mount
      upstream: anilist
      query: query {}
  - name: Mutation
    actions:
    - type: mount
      upstream: anilist
      query: mutation {}
```

### Importing a nested field of upstream graphql upstream.

Use a full graphql query to access nested child graph elements of the upstream
graphql server.  Feel free to use argument variables or literals in the query. 
variables used in the query will be translated into arguments for the newly defined
field. 

```yaml
types:
  - name: Query
    actions:
    - type: mount
      field: pagedCharacterSearch
      upstream: anilist
      query: |
        query ($search:String, $page:Int) {
          Page(page:$page, perPage:10) {
            characters(search:$search)
          }
        }
```

In the example above a `pagedCharacterSearch(search:String, page:Int)` field would 
be defined on the Query type and it would return the type returned by the `characters`
field of the anilist upstream. 

## License

[BSD](./LICENSE)

## Development

- We love [pull requests](https://github.com/chirino/graphql-gw/pulls)
- [Open Issues](https://github.com/chirino/graphql-gw/issues)
- graphql-gw is written in [Go](https://golang.org/). It should work on any platform where go is supported.
- Built on this [GraphQL](https://github.com/chirino/graphql) framework
