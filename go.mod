module github.com/aerogear/graphql-link

require (
	github.com/chirino/graphql v0.0.0-20200723175208-cec7bf430a98
	github.com/chirino/graphql-4-apis v0.0.0-20210703144953-6bdae7f3c2d0
	github.com/chirino/hawtgo v0.0.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/golang/protobuf v1.5.2
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.38.0
	gopkg.in/yaml.v2 v2.4.0
)

go 1.13

//replace github.com/chirino/graphql => ../graphql
//replace github.com/chirino/graphql-4-apis => ../graphql-4-apis
