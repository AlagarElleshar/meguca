// Simple proxy for cleaner directory structure and so we can have godoc support

package main

import "github.com/bakape/meguca/server"

//go:generate qtc -dir=templates -ext=html

func main() {
	err := server.Start()
	if err != nil {
		panic(err)
	}
}
