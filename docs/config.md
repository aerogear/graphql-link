## Configuration Guide

## Configuration Root Fields

| Field        | Required | Type                       | Description                                                                |
| ------------ | -------- | -------------------------- | -------------------------------------------------------------------------- |
| `listen:`    | no       | string                     | sets the host and port you want the graphql server to listen on.           |
| `upstreams:` | no       | [Upstream Map](#upstreams) | map of all the upstream severs that you will be accessing with the gateway |
| `schema:`    | no       | [Schema](#schema)          |                                                                            |
| `types:`     | no       | [Types](#types)            |                                                                            |

Example:

```yaml
listen: localhost:8080
upstreams:
  anilist:
    url: https://graphql.anilist.co/
types:
  - name: Query
    actions:
```

### Upstreams

The `upstreams:` section holds a map of all the upstream severs that you will be
accessing with the gateway.  The default upstream type is `graphql`. 

| type                            | Description                                                                       |
| ------------------------------- | --------------------------------------------------------------------------------- |
| [`graphql`](#action-type-mount) | the upstream server implements graphql                                            |
| [`openapi`](#action-type-mount) | the upstream server implements a REST interface described by an openapi document. |

### Upstream `type: graphql`

The `graphql` upstream type supports the following configuration options:

| Field      | Required | Type                  | Description                                                                                          |
| ---------- | -------- | --------------------- | ---------------------------------------------------------------------------------------------------- |
| `url:`     | yes      | url                   | the URL to the graphql endpoint                                                                      |
| `prefix:`  | no       | string                | A prefix to add to all upstream graphql Types when they get imported into the gateway graphql schema |
| `suffix:`  | no       | string                | A suffix to add to all upstream graphql Types when they get imported into the gateway graphql schema |
| `headers:` | no       | [Headers]((#headers)) | A Headers configuration section                                                                      |

### Upstream `type: openapi`

The `openapi` upstream type supports the following configuration options:

| Field      | Required | Type                    | Description                                                                                          |
| ---------- | -------- | ----------------------- | ---------------------------------------------------------------------------------------------------- |
| `spec:`    | yes      | [Endpoint]((#endpoint)) | Where the openapi specification document can be obtained.                                            |
| `api:`     | no       | [Endpoint]((#endpoint)) | Sets where endpoint base URL is accessed                                                             |
| `prefix:`  | no       | string                  | A prefix to add to all upstream graphql Types when they get imported into the gateway graphql schema |
| `suffix:`  | no       | string                  | A suffix to add to all upstream graphql Types when they get imported into the gateway graphql schema |
| `headers:` | no       | [Headers]((#headers))   | A Headers configuration section                                                                      |

### Headers

A headers section supports the following configuration options:

| Field                 | type                | Description                                                  |
| --------------------- | ------------------- | ------------------------------------------------------------ |
| `disable-forwarding:` | boolean             | disables forwarding client set headers to the upstream proxy |
| `set:`                | list of name values | Headers to set on the on the upstream reqeust                |
| `remove:`             | list of strings     | headers to remove the upstream request                       |

### Endpoint

An Endpoint configuration section supports the following configuration options:

| Field              | Required | Type   | Description                                                                               |
| ------------------ | -------- | ------ | ----------------------------------------------------------------------------------------- |
| `url:`             | yes      | url    | the URL to the endpoint                                                                   |
| `bearer-token:`    | yes      | string | an Authentication Bearer token that will added to the request headers.                    |
| `insecure-client:` | yes      | string | allows the client request to connect to TLS servers that do not have a valid certificate. |
| `api-key:`         | yes      | string | the API key to use with against the API (as defined in the openapi spec)                  |

### `schema:`

The optional schema section allows you configure the root query type names.  The default values of those fields are shown in the folllowing table.

| Field       | Default        |
| ----------- | -------------- |
| `query:`    | `Query`        |
| `mutation:` | `Mutation`     |
| `schema:`   | `Subscription` |

The example below only the root mutaiton type name is changed from it's default value to `GatewayMutation`.  Doing something like this can be useful if you do not want to rename type name from an upstream it contains a type name for one of these root query type names.

```yaml
schema:
  mutation: GatewayMutation
```

### `types:`

Use the `types:` section of the configuration to define the fields that can be 
accessed by clients of the `graphql-gw`.   You typicaly start by configuring the root query type names

The following example will add a field `myfield` to the `Query` type where the type is the root query of the `anilist` upstream server. 

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

`actions:` is a list configuration actions to take against on the named type.  The actions are processed in order.  You can select from the following actions types:

| type                            | Description                                                               |
| ------------------------------- | ------------------------------------------------------------------------- |
| [`mount`](#action-type-mount)   | mounts an upstream field onto a gateway schema type using a graphql query |
| [`rename`](#action-type-rename) | renames either a type or field in the gateway schema.                     |
| [`remove`](#action-type-remove) | used to remove a field from a type.                                       |
| [`link`](#action-type-link)     | Used to create graph links between types from different servers.          |

### Action `type: mount`

The `mount` action can be used to mount an upstream field onto a gateway schema type using a graphql query

| Field       | Required | Description                                                                                                                             |
| ----------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `upstream:` | yes      | a reference to an upstream server defined in the `upstreams:` section.                                                                  |
| `query:`    | yes      | partial graphql query document to one node in the upstream server graph.                                                                |
| `field:`    | no       | field name to mount the resulting node on to.  not not specified, then all the field of the node are mounted on to the the parent type. |

### Action `type: link`

The `link` action is a more advanced version of mount.  It is typically used to create new fields that link to data from a different upstream server.

| Field       | Required | Description                                                                                                                                      |
| ----------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `upstream:` | yes      | a reference to an upstream server defined in the `upstreams:` section.                                                                           |
| `query:`    | yes      | partial graphql query document to one node in the upstream server graph.                                                                         |
| `field:`    | yes      | field name to mount the resulting node on to.                                                                                                    |
| `vars:`     | no       | a map of variable names to query selection paths to single leaf node.  The selections defined here will be the values passed to the the `query:` |

### Action `type: rename`

The `rename` action can be used to rename either a type or field in the gateway schema. 

| Field    | Required | Description                                                                                  |
| -------- | -------- | -------------------------------------------------------------------------------------------- |
| `field:` | no       | if not set, you will be renaming the type, if set, you will be renaming a field of the type. |
| `to:`    | yes      | the new name                                                                                 |

### Action `type: remove`

The `remove` action can be used to remove a field from a type.

| Field    | Required | Description                            |
| -------- | -------- | -------------------------------------- |
| `field:` | yes      | The field name to remove from the type |

## Common Use Cases

### Importing all the fields of an upstream graphql server

If you want to import all the fields of an upstream server type, simply don't specify the name for the field to mount the query on to. The following example will import all the query fields onto the `Query` type and all the mutation fields on the `Mutation` type. 

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

### Importing a nested field of upstream graphql server

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

In the example above a `pagedCharacterSearch(search:String, page:Int)` field would be defined on the Query type and it would return the type returned by the `characters`field of the anilist upstream. 
