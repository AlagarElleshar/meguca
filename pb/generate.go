package pb

//go:generate protoc --go_out=. --go_opt=paths=source_relative messages.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative nekotvstate.proto
