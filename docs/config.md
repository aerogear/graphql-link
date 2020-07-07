## Configuration Guide

### `listen:`

Set the `listen:` to the host and port you want the graphql server to listen on. 

Example:

```yaml
listen: localhost:8080
```

### `upstreams:`

The upstreams section holds a map of all the upstream severs that you will be
accessing with the gateway.  The example below defines two upstream servers: `anilist` and `users`. Keep in mind that the URL configured must be a graphql server that is accessible from the gateway's network. 

```yaml
upstreams:
  anilist:
    url: https://graphql.anilist.co/
  users:
    url: https://users.acme.io/graphql
```

If there are duplicate types across the upstream severs you can configure either type name prefixes or suffixes on the upstream severs so that conflicts can be avoided when imported into the gateway.  Example:

```yaml
upstreams:
  anilist:
    url: https://graphql.anilist.co/
    prefix: Ani
    suffix: Type
  users:
    url: https://users.acme.io/graphql
```

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
