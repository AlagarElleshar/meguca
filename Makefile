export node_bins=$(PWD)/node_modules/.bin
export uglifyjs=$(node_bins)/uglifyjs
export gulp=$(node_bins)/gulp
export webpack=$(node_bins)/webpack
export GO111MODULE=on

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	ROCKSDB_CFLAGS := $(shell pkg-config --cflags-only-I rocksdb liblz4 libzstd) -I/opt/homebrew/opt/snappy/include
	ROCKSDB_LDFLAGS := $(shell pkg-config --libs-only-l --libs-only-L rocksdb liblz4 libzstd) -L/opt/homebrew/opt/snappy/lib
else ifeq ($(UNAME_S),Linux)
	ROCKSDB_CFLAGS := -I$(HOME)/rocksdb/include
	ROCKSDB_LDFLAGS := -L$(HOME)/rocksdb -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd -lbz2
endif

ifeq ($(UNAME_S),Linux)
    GO_BUILD_TAGS = -tags "libsqlite3 linux"
endif

.PHONY: client server imager test

all: client server

client: client_deps
	node esbuild.config.js --css --js

client_deps:
	npm install --include=dev --progress false --depth 0

css:
	node esbuild.config.js --css

js:
	node esbuild.config.js --js

generate:
	go generate ./...

server:
	go generate
	CGO_CFLAGS="$(ROCKSDB_CFLAGS)" CGO_LDFLAGS="$(ROCKSDB_LDFLAGS)" go build -v $(GO_BUILD_TAGS)

client_clean:
	rm -rf www/js www/css/*.css www/css/maps node_modules manifest.json

clean: client_clean
	rm -rf .build .ffmpeg .package target meguca-*.zip meguca-*.tar.xz meguca meguca.exe server/pkg

test:
	go test --race ./...

test_no_race:
	go test ./...

test_docker:
	docker-compose build
	docker-compose run --rm -e CI=true meguca make test
