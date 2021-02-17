(window.webpackJsonp=window.webpackJsonp||[]).push([[4],{73:function(e,t,n){"use strict";n.r(t),n.d(t,"frontMatter",(function(){return r})),n.d(t,"metadata",(function(){return o})),n.d(t,"toc",(function(){return s})),n.d(t,"default",(function(){return d}));var a=n(3),i=n(7),l=(n(0),n(88)),r={id:"installation",title:"Installation",sidebar_label:"Installation",slug:"/installation"},o={unversionedId:"installation",id:"installation",isDocsHomePage:!1,title:"Installation",description:"Installing Prebuilt Binaries",source:"@site/docs/installation.md",slug:"/installation",permalink:"/graphql-link/docs/installation",editUrl:"https://github.com/aerogear/graphql-link/docs/installation.md",version:"current",sidebar_label:"Installation",sidebar:"main",previous:{title:"Why use Graphql Link?",permalink:"/graphql-link/docs/"},next:{title:"CLI Guide",permalink:"/graphql-link/docs/cli-guide"}},s=[{value:"Installing Prebuilt Binaries",id:"installing-prebuilt-binaries",children:[]},{value:"Installing from Source",id:"installing-from-source",children:[]},{value:"Getting started",id:"getting-started",children:[{value:"Development and Production Mode",id:"development-and-production-mode",children:[]},{value:"Demos",id:"demos",children:[]}]},{value:"Guides",id:"guides",children:[]},{value:"Build from source",id:"build-from-source",children:[]},{value:"Docker image",id:"docker-image",children:[]}],c={toc:s};function d(e){var t=e.components,n=Object(i.a)(e,["components"]);return Object(l.b)("wrapper",Object(a.a)({},c,n,{components:t,mdxType:"MDXLayout"}),Object(l.b)("h3",{id:"installing-prebuilt-binaries"},"Installing Prebuilt Binaries"),Object(l.b)("p",null,"Please download ",Object(l.b)("a",{parentName:"p",href:"https://github.com/aerogear/graphql-link/releases"},"latest github release")," for your platform"),Object(l.b)("h3",{id:"installing-from-source"},"Installing from Source"),Object(l.b)("p",null,"If you have a recent ",Object(l.b)("a",{parentName:"p",href:"https://golang.org/dl/"},"go")," SDK installed:"),Object(l.b)("p",null,Object(l.b)("inlineCode",{parentName:"p"},"go get -u github.com/aerogear/graphql-link")),Object(l.b)("h2",{id:"getting-started"},"Getting started"),Object(l.b)("p",null,"Use the following command to create a default server configuration file."),Object(l.b)("pre",null,Object(l.b)("code",{parentName:"pre",className:"language-bash"},"$ graphql-link config init\n\nCreated:  graphql-link.yaml\n\nStart the gateway by running:\n\n    graphql-link serve\n\n")),Object(l.b)("p",null,"Then run the server using this command:"),Object(l.b)("pre",null,Object(l.b)("code",{parentName:"pre",className:"language-bash"},"$ graphql-link serve\n2020/07/07 10:16:29 GraphQL endpoint is running at http://127.0.0.1:8080/graphql\n2020/07/07 10:16:29 Gateway Admin UI and GraphQL IDE is running at http://127.0.0.1:8080\n")),Object(l.b)("p",null,"You can then use the Web UI at ",Object(l.b)("a",{parentName:"p",href:"http://127.0.0.1:8080"},"http://127.0.0.1:8080")," to configure the gateway."),Object(l.b)("h3",{id:"development-and-production-mode"},"Development and Production Mode"),Object(l.b)("p",null,"The ",Object(l.b)("inlineCode",{parentName:"p"},"graphql-link serve")," command will run the gateway in development mode. Development mode enables the configuration web interface and will cause the gateway to periodical download upstream schemas on start up. The schema files will be stored in the ",Object(l.b)("inlineCode",{parentName:"p"},"upstreams")," directory (located in the same directory as the gateway configuration file). If any of the schemas cannot be downloaded the gateway will fail to startup."),Object(l.b)("p",null,"You can use ",Object(l.b)("inlineCode",{parentName:"p"},"graphql-link serve --production")," to enabled production mode. In this mode, the configuration web interface is disabled, and the schema for the upstream severs will be loaded from the ",Object(l.b)("inlineCode",{parentName:"p"},"upstreams")," directory that they were stored when you used development mode. This ensures that your gateway will have a consistent schema presented, and that it's start up will not be impacted by the availability of the upstream\nservers."),Object(l.b)("h3",{id:"demos"},"Demos"),Object(l.b)("ul",null,Object(l.b)("li",{parentName:"ul"},Object(l.b)("a",{parentName:"li",href:"https://www.youtube.com/watch?v=I5AStj2csD0"},"https://www.youtube.com/watch?v=I5AStj2csD0"))),Object(l.b)("h2",{id:"guides"},"Guides"),Object(l.b)("ul",null,Object(l.b)("li",{parentName:"ul"},Object(l.b)("a",{parentName:"li",href:"/graphql-link/docs/config"},"Yaml Configuration Guide")),Object(l.b)("li",{parentName:"ul"},Object(l.b)("a",{parentName:"li",href:"/graphql-link/docs/cli-guide"},"CLI Guide"))),Object(l.b)("h2",{id:"build-from-source"},"Build from source"),Object(l.b)("pre",null,Object(l.b)("code",{parentName:"pre",className:"language-bash"},"go build -o=graphql-link main.go\n")),Object(l.b)("h2",{id:"docker-image"},"Docker image"),Object(l.b)("pre",null,Object(l.b)("code",{parentName:"pre"},"docker pull aerogear/graphql-link\n")))}d.isMDXComponent=!0}}]);