package main

import "github.com/marcell7/domena-check/server"

func main() {
	server := server.New()
	server.Start()
}
