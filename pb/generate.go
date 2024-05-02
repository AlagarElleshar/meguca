package pb

//go:generate protoc --go_out=. --go_opt=paths=source_relative --plugin=protoc-gen-ts=protoc-gen-ts_proto --ts_proto_out=../client/typings messages.proto
