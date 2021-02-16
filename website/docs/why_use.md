---
id: why_use
title: Why use Graphql Link?
sidebar_label: Why use Graphql Link
slug: /
---

> GraphQL-Link is an GraphQL gateway that lets you easily proxy to other GraphQL servers.

<img src="https://raw.githubusercontent.com/aerogear/graphql-link/master/docs/images/logo.png" alt="logo-graphql-link"  />

### Features

- Consolidate access to multiple upstream GraphQL servers via a single GraphQL gateway server.
- Introspection of the upstream server to discover their GraphQL schemas.
- The configuration uses GraphQL queries to define which upstream fields and types can be accessed.
- Upstream types, that are accessible, are automatically merged into the gateway schema.
- Type conflict due to the same type name existing in multiple upstream servers can be avoided by renaming types in the gateway.
- Supports GraphQL Queries, Mutations, and Subscriptions
- Production mode settings to avoid the gateway's schema from dynamically changing due to changes in the upstream schemas.
- Uses the dataloader pattern to batch multiple query requests to the upstream servers.
- Link the graphs of different upstream servers by defining additional link fields.
- Web based configuration UI
- OpenAPI based upstream servers (get automatically converted to a GraphQL Schema)

<img src="https://raw.githubusercontent.com/aerogear/graphql-link/master/docs/images/graphql-link-overview.jpg" alt="diagram of graphql-link" />
