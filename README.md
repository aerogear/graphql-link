# graphql-link

<img src="https://raw.githubusercontent.com/aerogear/graphql-link/master/docs/images/logo/png" alt="logo-graphql-link" width="488">

This is a GraphQL gateway that lets you easily compose or federate other
GraphQL upstream servers.

<img src="https://raw.githubusercontent.com/aerogear/graphql-link/master/docs/images/graphql-link-overview.jpg" alt="diagram of graphql-link" width="488">

## Features

* Consolidate access to multiple upstream GraphQL servers via a single GraphQL gateway server.
* Introspection of the upstream server to discover their GraphQL schemas.
* The configuration uses GraphQL queries to define which upstream fields and types can be accessed.    
* Upstream types, that are accessible, are automatically merged into the gateway schema.
* Type conflict due to the same type name existing in multiple upstream servers can be avoided by renaming types in the gateway.
* Supports GraphQL Queries, Mutations, and Subscriptions
* Production mode settings to avoid the gateway's schema from dynamically changing due to changes in the upstream schemas.  
* Uses the dataloader pattern to batch multiple query requests to the upstream servers.
* Link the graphs of different upstream servers by defining additional link fields.
* Web based configuration UI
* OpenAPI based upstream servers (get automatically converted to a GraphQL Schema)

### Installing Prebuilt Binaries

Please download [latest github release](https://github.com/aerogear/graphql-link/releases) for your platform

### Installing from Source

If you have a recent [go](https://golang.org/dl/) SDK installed:

`go get -u github.com/aerogear/graphql-link`

## Getting started

Use the following command to create a default server configuration file.

```bash
$ graphql-link config init

Created:  graphql-link.yaml

Start the gateway by running:

    graphql-link serve

```

Then run the server using this command:

```bash
$ graphql-link serve
2020/07/07 10:16:29 GraphQL endpoint is running at http://127.0.0.1:8080/graphql
2020/07/07 10:16:29 Gateway Admin UI and GraphQL IDE is running at http://127.0.0.1:8080
```

You can then use the Web UI at [http://127.0.0.1:8080](http://127.0.0.1:8080) to configure the gateway.
 
### Development and Production Mode

The `graphql-link serve` command will run the gateway in development mode.  Development mode enables the configuration web interface and will cause the gateway to periodical download upstream schemas on start up.  The schema files will be stored in the `upstreams` directory (located in the same directory as the gateway configuration file).  If any of the schemas cannot be downloaded the gateway will fail to startup.

You can use `graphql-link serve --production` to enabled production mode.  In this mode, the configuration web interface is disabled, and the schema for the upstream severs will be loaded from the `upstreams` directory that they were stored when you used development mode.  This ensures that your gateway will have a consistent schema presented, and that it's start up will not be impacted by the availability of the upstream
servers.

### Demos

* https://www.youtube.com/watch?v=I5AStj2csD0

## Guides

* [Yaml Configuration Guide](docs/config.md)
* [CLI Guide](docs/cli.md)
 
## Build from source

```bash
go build -o=graphql-link main.go
```
## Docker image

```
docker pull aerogear/graphql-link
```

## License

[BSD](./LICENSE)

## Development

- We love [pull requests](https://github.com/aerogear/graphql-link/pulls)
- [Open Issues](https://github.com/aerogear/graphql-link/issues)
- graphql-link is written in [Go](https://golang.org/). It should work on any platform where go is supported.
- Built on this [GraphQL](https://github.com/chirino/graphql) framework

## History
 
Project was initialy build by @chirino as `graphql-gw`. 