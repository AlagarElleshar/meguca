module github.com/bakape/meguca

go 1.21

toolchain go1.22.2

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.0

replace github.com/bakape/thumbnailer/v2 => github.com/meowmin/thumbnailer/v2 v2.0.5

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/aquilax/tripcode v1.0.1
	github.com/aws/aws-sdk-go v1.53.3
	github.com/badoux/goscraper v0.0.0-20190827161153-36995ce6b19f
	github.com/bakape/captchouli/v2 v2.2.2
	github.com/bakape/thumbnailer/v2 v2.7.1
	github.com/chai2010/webp v1.1.1
	github.com/dimfeld/httptreemux v5.0.1+incompatible
	github.com/facebookgo/grace v0.0.0-20180706040059-75cf19382434
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-playground/log v6.3.0+incompatible
	github.com/google/generative-ai-go v0.12.0
	github.com/gorilla/websocket v1.5.1
	github.com/lib/pq v1.10.9
	github.com/linxGnu/grocksdb v1.9.1
	github.com/oschwald/geoip2-golang v1.9.0
	github.com/rakyll/statik v0.1.7
	github.com/rivo/uniseg v0.4.7
	github.com/ulikunitz/xz v0.5.12
	github.com/valyala/quicktemplate v1.7.0
	golang.org/x/crypto v0.23.0
	golang.org/x/text v0.15.0
	google.golang.org/api v0.181.0
	google.golang.org/protobuf v1.34.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/vansante/go-ffprobe.v2 v2.1.1
)

require (
	cloud.google.com/go v0.113.0 // indirect
	cloud.google.com/go/ai v0.5.0 // indirect
	cloud.google.com/go/auth v0.4.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.2 // indirect
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	cloud.google.com/go/longrunning v0.5.7 // indirect
	github.com/bakape/boorufetch v1.1.6 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/freeport v0.0.0-20150612182905-d4adf43b75b9 // indirect
	github.com/facebookgo/httpdown v0.0.0-20180706035922-5979d39b15c2 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/stats v0.0.0-20151006221625-1b76add642e4 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/ansi v2.1.0+incompatible // indirect
	github.com/go-playground/errors v3.3.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.4 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/oschwald/maxminddb-golang v1.12.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	gitlab.com/nyarla/go-crypt v0.0.0-20160106005555-d9a5dc2b789b // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.51.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.51.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240506185236-b8a5c65736ae // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/grpc v1.63.2 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)
