package main

import "github.com/tanisobe/ymp3d/ymp3d"

func main() {
	server := ymp3d.NewServer()
	server.Run()
}
