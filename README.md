# graphql-gw

This is a GraphQL gateway that lets you easily compose or federate other
GraphQL endpoints.

### Installing

`go get -U github.com/chirino/graphql-gw`

## Getting started

Run

```
graphql-gw new myproject
cd myproject
# edit myproject/graphql-gw.yaml to configure
# the gateway
graphql-gw serve
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

## License

[BSD](./LICENSE)

## Development

- We love [pull requests](https://github.com/chirino/graphql-gw/pulls)
- [Open Issues](https://github.com/chirino/graphql-gw/issues)
- graphql-gw is written in [Go](https://golang.org/). It should work on any platform where go is supported.
- Built on this [GraphQL](https://github.com/chirino/graphql) framework
