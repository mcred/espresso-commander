package main

import (
    "encoding/json"
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

// CommandRequest struct for incoming request
type CommandRequest struct {
    Type    string `json:"type"`    // "ping" or "sysinfo"
    Payload string `json:"payload"` // For ping, this is the host
}

// CommandResponse struct for outgoing resposne
type CommandResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Error   string      `json:"error,omitempty"`
}

func handleCommand(cmdr Commander) http.HandlerFunc {
    return middleware(func(w http.ResponseWriter, r *http.Request) {
        // get request struct from body
        var req CommandRequest
        body := json.NewDecoder(r.Body)
        err := body.Decode(&req)
        if err != nil {
            panic(err)
        }
        defer r.Body.Close()

        // prepare response struct
        var res CommandResponse
        switch req.Type {
        case "ping":
            p, err := cmdr.Ping(req.Payload)
            if err != nil {
                panic(err)
            }
            res.Success = p.Successful
            res.Data = p.Time
            break
        case "sysinfo":
            s, err := cmdr.GetSystemInfo()
            if err != nil {
                panic(err)
            }
            res.Success = true
            res.Data = s
            break
        default:
            panic("invalid request type")
        }

        // encode and send response
        w.Header().Set("Content-Type", "application/json")
        resJSON, err := json.Marshal(res)
        if err != nil {
            panic(err)
        }
        _, err = w.Write(resJSON)
        if err != nil {
            panic(err)
        }
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
