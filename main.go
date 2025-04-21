package main

import (
	"flag"
	"fmt"
	"harmony/benchmark"
	"harmony/cli"
	"harmony/handler"
	"log"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(0)

	cli.StartCli()
}

func startWebsocket() {
	slog.Info("Starting websocket...")

	// settings for the http service
	addr := flag.String("addr", "0.0.0.0:8080", "http service address")

	// define entrypoint as well as the callback function which will handle requests
	http.HandleFunc("/", handler.Serve)
	http.HandleFunc("/hello", HelloServer)

	// start listening for incoming connections
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func startBenchmark() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	go startWebsocket()
	time.Sleep(1 * time.Second)

	benchmark.StartBenchmark("ws://0.0.0.0:8080", 1000, 10000)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hi mom!")
}
