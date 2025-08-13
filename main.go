package main

import (
    "fmt"
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

type CommandRequest struct {
    Type    string `json:"type"`    // "ping" or "sysinfo"
    Payload string `json:"payload"` // For ping, this is the host
}

type CommandResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Error   string      `json:"error,omitempty"`
}

func handleCommand(cmdr Commander) http.HandlerFunc {
    return middleware(func(w http.ResponseWriter, r *http.Request) {
        log.Println(r.Method, r.URL.Path)
        panic("testing")
    })
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var response CommandResponse
        defer func() {
            // Catch all Panics and throw a 500
            if r := recover(); r != nil {
                log.Printf("Recovered from panic: %v\n", r)
                response.Success = false
                response.Error = fmt.Sprintf("%v", r)
                w.WriteHeader(http.StatusInternalServerError)
            }
        }()
        // disallow paths other than /execute
        if r.URL.Path != "/execute" {
            response.Success = false
            response.Error = "invalid path"
            w.WriteHeader(http.StatusNotFound)
        }
        // disallow methods other than POST
        if r.Method != "POST" {
            response.Success = false
            response.Error = "invalid method"
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
        // pass through to handler
        next(w, r)
    }
}
