package main

import (
    "log"
    "net/http"
)

func main() {
    commander := NewCommander()
    server := &http.Server{
        Addr:    ":8080",
        Handler: handleRequests(commander),
    }
    log.Fatal(server.ListenAndServe())
}

func handleRequests(cmdr Commander) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/execute", handleCommand(cmdr))
    return mux
}

func handleCommand(cmdr Commander) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse request and execute command
    }
}
