import React from "react";
import Layout from "@theme/Layout";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Body from "../components/homeBody";
import Hero from "../components/hero/hero";
// import "tailwindcss/tailwind.css";

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig = {} } = context;
  return (
    <Layout
      title={`Hello from ${siteConfig.title}`}
      description="Description will go into a meta tag in <head />"
    >
      <Hero />
      <main>
        <Body />
      </main>
    </Layout>
  );
}

export default Home;
