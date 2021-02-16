import React from "react";
import clsx from "clsx";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import useBaseUrl from "@docusaurus/useBaseUrl";
import styles from "./styles.module.css";

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

function Feature({ title, description }) {
  return (
    <div className={clsx("col col--4", styles.feature)}>
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig = {} } = context;
  return (
    <Layout
      title={`Hello from ${siteConfig.title}`}
      description="Description will go into a meta tag in <head />"
    >
      <header className={clsx("hero hero--primary", styles.heroBanner)}>
        <div className="container">
          <h1 className="hero__title">{siteConfig.title}</h1>
          <p className="hero__subtitle">{siteConfig.tagline}</p>
          <div className={styles.buttons}>
            <Link
              className={clsx(
                "button  button--outline button--secondary button--lg",
                styles.getStarted
              )}
              to={useBaseUrl("docs/")}
            >
              Get Started
            </Link>
          </div>
        </div>
      </header>
      <main>
        {features && features.length > 0 && (
          <section className={styles.features}>
            <div className="container">
              <div className="row">
                {features.map((props, idx) => (
                  <Feature key={idx} {...props} />
                ))}
              </div>
            </div>
          </section>
        )}
      </main>
    </Layout>
  );
}

export default Home;
