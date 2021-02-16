module.exports = {
  title: "Graphql Link",
  tagline:
    "Graphql Link is a GraphQL server that composes other GraphQL or OpenAPI endpoints",
  url: "https://aerogear.github.io",
  baseUrl: "/graphql-link/",
  onBrokenLinks: "ignore",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/logo.ico",
  organizationName: "aerogear",
  projectName: "graphql-link",
  plugins: ["docusaurus-tailwindcss-loader"],
  themeConfig: {
    navbar: {
      title: "Graphql Link",
      // logo: {
      //   alt: "graphql link",
      //   src: "img/logo.png",
      // },
      items: [
        {
          to: "docs/",
          activeBasePath: "docs",
          label: "Docs",
          position: "left",
        },
        {
          href: "https://github.com/aerogear/graphql-link/",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Docs",
          items: [
            {
              label: "CLI",
              to: "docs/",
            },
            {
              label: "Configuration",
              to: "docs/config/",
            },
          ],
        },
        {
          title: "Community",
          items: [
            {
              label: "Discord",
              href: "https://discord.gg/tfQ9jSzs9D",
            },
          ],
        },
        {
          title: "More",
          items: [
            {
              label: "GitHub",
              href: "https://github.com/aerogear/graphql-link/",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Aerogear`,
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/aerogear/graphql-link/",
        },
        blog: {
          showReadingTime: true,
          editUrl: "https://github.com/aerogear/graphql-link/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],
};
