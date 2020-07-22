module github.com/chirino/graphql-gw

require (
	github.com/chirino/graphql v0.0.0-20200723150749-b1992f46a318
	github.com/chirino/graphql-4-apis v0.0.0-20200622132210-66059999b694
	github.com/chirino/hawtgo v0.0.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/getkin/kin-openapi v0.14.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/golang/protobuf v1.3.1
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/sys v0.0.0-20200620081246-981b61492c35 // indirect
	google.golang.org/grpc v1.21.0
	gopkg.in/yaml.v2 v2.3.0
)

go 1.13

//replace github.com/chirino/graphql => ../graphql
// replace github.com/chirino/graphql => ../graphql

//replace github.com/chirino/graphql-4-apis => ../graphql-4-apis
