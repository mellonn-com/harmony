package main

import (
	"flag"
	"harmony/benchmark"
	"harmony/handler"
	"log"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(0)

	startBenchmark()
	// startWebsocket()
}

func startWebsocket() {
	slog.Info("Starting websocket...")

	// settings for the http service
	addr := flag.String("addr", "0.0.0.0:8080", "http service address")

	// define entrypoint as well as the callback function which will handle requests
	http.HandleFunc("/", handler.Serve)

	// start listening for incoming connections
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func startBenchmark() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	go startWebsocket()
	time.Sleep(1 * time.Second)

	benchmark.StartBenchmark("ws://0.0.0.0:8080", 110, 10000)
}
