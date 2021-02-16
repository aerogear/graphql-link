---
id: installation
title: Installation
sidebar_label: Installation
slug: /installation
---

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

The `graphql-link serve` command will run the gateway in development mode. Development mode enables the configuration web interface and will cause the gateway to periodical download upstream schemas on start up. The schema files will be stored in the `upstreams` directory (located in the same directory as the gateway configuration file). If any of the schemas cannot be downloaded the gateway will fail to startup.

You can use `graphql-link serve --production` to enabled production mode. In this mode, the configuration web interface is disabled, and the schema for the upstream severs will be loaded from the `upstreams` directory that they were stored when you used development mode. This ensures that your gateway will have a consistent schema presented, and that it's start up will not be impacted by the availability of the upstream
servers.

### Demos

- https://www.youtube.com/watch?v=I5AStj2csD0

## Guides

- [Yaml Configuration Guide](config.md)
- [CLI Guide](cli.md)

## Build from source

```bash
go build -o=graphql-link main.go
```

## Docker image

```
docker pull aerogear/graphql-link
```
