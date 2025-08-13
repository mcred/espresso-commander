package main

import (
    "os"
    "time"
)

type Commander interface {
    Ping(host string) (PingResult, error)
    GetSystemInfo() (SystemInfo, error)
}

type PingResult struct {
    Successful bool
    Time       time.Duration
}

type SystemInfo struct {
    Hostname  string
    IPAddress string
}
type commander struct{}

func (c *commander) Ping(host string) (PingResult, error) {
    //TODO implement me
    panic("implement me")
}

func NewCommander() Commander {
    return &commander{}
}

func (c *commander) GetSystemInfo() (SystemInfo, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return SystemInfo{}, err
    }

    // Get IP address (implement this)

    return SystemInfo{
        Hostname:  hostname,
        IPAddress: "implement me",
    }, nil
}
