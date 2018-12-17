package main

import (
	"github.com/whomm/hrproxy/httptool"
)

func main() {
	serv := httptool.NewServer(":8091", "http://127.0.0.1:9200")
	serv.Handler(`^.*/_cat/?.*$`, 2, nil)
	serv.ListenAndServe()
}
