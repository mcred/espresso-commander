package main

import (
    "fmt"
    probing "github.com/prometheus-community/pro-bing"
    "log"
    "net"
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
    // built from examples in
    // https://github.com/prometheus-community/pro-bing
    s := false
    t := time.Duration(0)

    pinger, err := probing.NewPinger(host)
    if err != nil {
        return PingResult{}, err
    }

    pinger.OnRecv = func(pkt *probing.Packet) {
        log.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
            pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.TTL)
    }
    pinger.OnDuplicateRecv = func(pkt *probing.Packet) {
        log.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
            pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.TTL)
    }
    pinger.OnFinish = func(stats *probing.Statistics) {
        log.Printf("\n--- %s ping statistics ---\n", stats.Addr)
        log.Printf("%d packets transmitted, %d packets received, %d duplicates, %v%% packet loss\n",
            stats.PacketsSent, stats.PacketsRecv, stats.PacketsRecvDuplicates, stats.PacketLoss)
        log.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
            stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
        s = true
        t = stats.MaxRtt
    }

    pinger.Count = 4
    pinger.Size = 24
    pinger.Interval = time.Second
    pinger.Timeout = time.Second * 100000
    pinger.TTL = 64
    pinger.SetPrivileged(false)

    log.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
    err = pinger.Run()
    if err != nil {
        panic(fmt.Errorf("Failed to ping target host:", err))
    }
    return PingResult{Successful: s, Time: t}, nil
}

func NewCommander() Commander {
    return &commander{}
}

func (c *commander) GetSystemInfo() (SystemInfo, error) {
    // Get the system hostname
    hostname, err := os.Hostname()
    if err != nil {
        return SystemInfo{}, err
    }

    // Initialize IP address variable
    ipAddress := ""
    
    // Get all network interfaces
    interfaces, err := net.Interfaces()
    if err != nil {
        return SystemInfo{}, err
    }
    
    // Iterate through interfaces to find an active, non-loopback interface
    for _, iface := range interfaces {
        // Skip interfaces that are down or loopback
        if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
            continue
        }
        
        // Get addresses for this interface
        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }
        
        // Check each address
        for _, addr := range addrs {
            var ip net.IP
            // Extract IP from address based on type
            switch v := addr.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            
            // Use the first IPv4 address found
            if ip != nil && ip.To4() != nil {
                ipAddress = ip.String()
                break
            }
        }
        
        // Stop searching if we found an IP
        if ipAddress != "" {
            break
        }
    }

    // Fallback to localhost if no suitable IP found
    if ipAddress == "" {
        ipAddress = "127.0.0.1"
    }

    return SystemInfo{
        Hostname:  hostname,
        IPAddress: ipAddress,
    }, nil
}
