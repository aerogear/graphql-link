---
id: cli
title: CLI Guide
sidebar_label: CLI Guide
slug: /cli-guide
---

### Create a Default Configuration

```bash
$ graphql-link config init

Created:  graphql-link.yaml

Start the gateway by running:

    graphql-link serve
```

### Run the Server

```bash
$ graphql-link serve
2021/02/16 18:46:48 GraphQL endpoint is running at http://127.0.0.1:8080/graphql
2021/02/16 18:46:48 Gateway Admin UI and GraphQL IDE is running at http://127.0.0.1:8080
```

### Add Upstream

```bash
$ graphql-link  config add upstream <NAME_OF_YOUR_UPSTEAM_SERVER> <URL_TO_GRAPHQL_ENDPOINT> --prefix <PREFIX_NAME_TO_APPLY_TO_ALL_UPSTREAM_SCHEMA_TYPES>

upstream added
```

### Mount Upstream Server Query Fields on to the Gateway Server

```bash
$ graphql-link config mount <NAME_OF_YOUR_UPSTEAM_SERVER> Query --field <FIELD_NAME>

mount added
```

### Link Entities of Different Upstream Servers

> Example of Adding a Pokemon Field to AnimeList Characters. This field will query the Pokemon service for the Pokemon where the name matches the anime's character's name

```bash
# Adding pokemon upstream
$ graphql-link config add upstream pokemon <POKEMON_GRAPHQL_URL> --prefix Pokemon
upstream added

# Mounting pokemon query fields
$ graphql-link config mount pokemon Query --field pokemonV1
mount added

# Adding anime upstream
$ graphql-link config add upstream amime <ANIME_GRAPHQL_URL> --prefix Anime
loaded previously stored schema: upstreams/pokemon.graphql
upstream added

# Mounting anime query fields
$ graphql-link config mount anime Query --field animeV1
loaded previously stored schema: upstreams/pokemon.graphql
loaded previously stored schema: upstreams/anime.graphql
mount added

# Linking the pokemon and anime entities
$ graphql-link config link pokemon AnimeCharacter pokemon --var '$n=name{full}' --query '{pokemon(name:$n)}'

loaded previously stored schema: upstreams/pokemon.graphql
loaded previously stored schema: upstreams/anime.graphql
link added
```

### Usage

`$ graphql-link --help`

```

A GraphQL composition gateway

Usage:
graphql-link [command]

Available Commands:
completion Generates bash completion scripts
config Modifies the gateway configuration
help Help about any command
serve Runs the gateway service
version Print version information for this executable

Flags:
-h, --help help for graphql-link
--verbose enables increased verbosity

Use "graphql-link [command] --help" for more information about a command.

```
