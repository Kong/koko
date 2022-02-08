module github.com/kong/koko

go 1.17

require (
	github.com/bufbuild/buf v0.56.0
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/gavv/httpexpect/v2 v2.3.1
	github.com/ghodss/yaml v1.0.0
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/hbagdi/gang v0.1.0
	github.com/imdario/mergo v0.3.12
	github.com/jeremywohl/flatten v1.0.1
	github.com/kong/go-kong v0.25.1
	github.com/kong/go-wrpc v0.0.0-20210914213024-d4348db6b815
	github.com/kong/goks v0.2.0
	github.com/lib/pq v1.10.3
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.13.0
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9
	go.uber.org/zap v1.19.1
	google.golang.org/genproto v0.0.0-20211026145609-4688e4c4e024
	google.golang.org/grpc v1.41.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/structs v1.0.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/imkira/go-interpol v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jhump/protoreflect v1.9.1-0.20210817181203-db1a327a393e // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/layeh/gopher-json v0.0.0-20201124131017-552bb3c4c3bf // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pkg/profile v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/twitchtv/twirp v8.1.0+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.27.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.1.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yuin/gluare v0.0.0-20170607022532-d7c94f1a80ed // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/net v0.0.0-20211013171255-e13a2654a71e // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	moul.io/http2curl v1.0.1-0.20190925090545-5cd742060b0e // indirect
)

replace github.com/yuin/gopher-lua => github.com/hbagdi/gopher-lua v0.0.0-20211129210354-3e4a277fb892

replace github.com/layeh/gopher-json => github.com/hbagdi/gopher-json v0.0.0-20211203171840-04cc2cd39713

replace github.com/jeremywohl/flatten => github.com/hbagdi/flatten v1.0.2-0.20211207185041-fe643c674d12
