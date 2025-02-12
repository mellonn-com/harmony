package main

import (
	"flag"
	"harmony/handler"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(0)
	log.Println("Starting websocket...")

	// settings for the http service
	addr := flag.String("addr", "0.0.0.0:8080", "http service address")

	// define entrypoint as well as the callback function which will handle requests
	http.HandleFunc("/", handler.Serve)

	// start listening for incoming connections
	log.Fatal(http.ListenAndServe(*addr, nil))
}
