import React from "react";
import clsx from "clsx";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import useBaseUrl from "@docusaurus/useBaseUrl";
import styles from "./styles.module.css";
import "tailwindcss/tailwind.css";

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

function Body() {
  return (
    <div>
      <div className="flex flex-wrap p-5">
        {features && features.length > 0 && (
          <div className="row p-10">
            {features.map((props, idx) => (
              <div
                key={idx}
                className="xl:w-1/3 lg:w-1/2 md:w-full px-8 py-6 my-5 border-l-2 border-gray-200 border-opacity-60"
              >
                <h3 className="text-lg sm:text-xl  font-medium title-font mb-2">
                  {props.title}
                </h3>
                <p className="leading-relaxed text-base mb-4">
                  {props.description}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>
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
                "mt-10 text-black px-10 py-3 rounded-md no-underline hover:bg-gray-200 hover:no-underline  bg-white",
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
        {/* {features && features.length > 0 && (
          <section className={styles.features}>
            <div className="container ">
              <div className="row">
                {features.map((props, idx) => (
                  <Feature key={idx} {...props} />
                ))}
              </div>
            </div>
          </section>
        )} */}

        <Body />
      </main>
    </Layout>
  );
}

export default Home;
