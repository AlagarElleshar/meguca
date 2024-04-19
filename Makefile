export node_bins=$(PWD)/node_modules/.bin
export uglifyjs=$(node_bins)/uglifyjs
export gulp=$(node_bins)/gulp
export webpack=$(node_bins)/webpack
export GO111MODULE=on

ROCKSDB_CFLAGS := $(shell pkg-config --cflags rocksdb liblz4 libzstd) -I/opt/homebrew/Cellar/snappy/1.1.10/include/
ROCKSDB_LDFLAGS := -lstdc++ -lm -lz -lsnappy -llz4 -lzstd
ROCKSDB_LDFLAGS += $(shell pkg-config --libs rocksdb liblz4 libzstd) -L/opt/homebrew/Cellar/snappy/1.1.10/lib

ifeq ($(shell uname -s),Linux)
    GO_BUILD_TAGS = -tags "libsqlite3 linux"
endif

.PHONY: client server imager test

all: client server

client: client_vendor
	$(webpack)
	$(gulp)

client_deps:
	npm install --include=dev --progress false --depth 0

client_vendor: client_deps
	mkdir -p www/js/vendor

css:
	$(gulp) css

generate:
	go generate ./...

server:
	go generate
	CGO_CFLAGS="$(ROCKSDB_CFLAGS)" CGO_LDFLAGS="$(ROCKSDB_LDFLAGS)" go build -v $(GO_BUILD_TAGS)

client_clean:
	rm -rf www/js www/css/*.css www/css/maps node_modules

clean: client_clean
	rm -rf .build .ffmpeg .package target meguca-*.zip meguca-*.tar.xz meguca meguca.exe server/pkg

test:
	go test --race ./...

test_no_race:
	go test ./...

test_docker:
	docker-compose build
	docker-compose run --rm -e CI=true meguca make test