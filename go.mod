module github.com/bakape/meguca

go 1.23.0

toolchain go1.24.2

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.0

replace github.com/bakape/thumbnailer/v2 => github.com/AlagarElleshar/thumbnailer/v2 v2.0.9

//replace github.com/bakape/thumbnailer/v2 => ../go-local/thumbnailer

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/aquilax/tripcode v1.0.1
	github.com/aws/aws-sdk-go-v2 v1.40.1
	github.com/aws/aws-sdk-go-v2/config v1.32.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.93.0
	github.com/badoux/goscraper v0.0.0-20190827161153-36995ce6b19f
	github.com/bakape/captchouli/v2 v2.2.2
	github.com/bakape/thumbnailer/v2 v2.7.1
	github.com/dimfeld/httptreemux v5.0.1+incompatible
	github.com/facebookgo/grace v0.0.0-20180706040059-75cf19382434
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-playground/log v6.3.0+incompatible
	github.com/gorilla/websocket v1.5.3
	github.com/lib/pq v1.10.9
	github.com/linxGnu/grocksdb v1.10.0
	github.com/oschwald/geoip2-golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/rivo/uniseg v0.4.7
	github.com/ulikunitz/xz v0.5.12
	github.com/valyala/quicktemplate v1.8.0
	golang.org/x/crypto v0.38.0
	golang.org/x/text v0.25.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/vansante/go-ffprobe.v2 v2.2.1
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.3 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/bakape/boorufetch v1.1.6 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/freeport v0.0.0-20150612182905-d4adf43b75b9 // indirect
	github.com/facebookgo/httpdown v0.0.0-20180706035922-5979d39b15c2 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/stats v0.0.0-20151006221625-1b76add642e4 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/go-playground/ansi v2.1.0+incompatible // indirect
	github.com/go-playground/errors v3.3.0+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/oschwald/maxminddb-golang v1.13.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	gitlab.com/nyarla/go-crypt v0.0.0-20160106005555-d9a5dc2b789b // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)
