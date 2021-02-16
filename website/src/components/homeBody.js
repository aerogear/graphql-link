import React from "react";

const features = [
  {
    title: "Multiple Upstreams",
    description: (
      <>
        Consolidate access to multiple upstream GraphQL servers via a single
        GraphQL gateway server. Introspection of the upstream server to discover
        their GraphQL schemas.
      </>
    ),
  },
  {
    title: "Configuration",
    description: (
      <>
        The configuration uses GraphQL queries to define which upstream fields
        and types can be accessed. Upstream types, that are accessible, are
        automatically merged into the gateway schema.
      </>
    ),
  },
  {
    title: "Avoid type conflicts",
    description: (
      <>
        Type conflict due to the same type name existing in multiple upstream
        servers can be avoided by renaming types in the gateway. Supports
        GraphQL Queries, Mutations, and Subscriptions
      </>
    ),
  },
  {
    title: "Dataloader pattern",
    description: (
      <>
        Production mode settings to avoid the gateway's schema from dynamically
        changing due to changes in the upstream schemas. Uses the dataloader
        pattern to batch multiple query requests to the upstream servers.
      </>
    ),
  },

  {
    title: "OpenAPI support",
    description: (
      <>
        Link the graphs of different upstream servers by defining additional
        link fields. Web based configuration UI OpenAPI based upstream servers
        (get automatically converted to a GraphQL Schema)
      </>
    ),
  },
];

export default function homeBody() {
  return (
    <div>
      <div class="row homebody">
        {features.map((props, idx) => (
          <div key={idx} class="col">
            <div class="card">
              <div class="card__header">
                <h3>{props.title}</h3>
              </div>
              <div class="card__body">
                <p>{props.description}</p>
              </div>
              <div class="card__footer"></div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
