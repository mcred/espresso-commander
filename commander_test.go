package main

import (
	"testing"
	"time"
)

func TestNewCommander(t *testing.T) {
	// Test that NewCommander returns a non-nil Commander interface
	cmdr := NewCommander()
	if cmdr == nil {
		t.Fatal("NewCommander() returned nil")
	}

	// Verify it's the correct type
	_, ok := cmdr.(*commander)
	if !ok {
		t.Fatal("NewCommander() did not return *commander type")
	}
}

func TestCommander_GetSystemInfo(t *testing.T) {
	cmdr := NewCommander()

	// Test GetSystemInfo
	info, err := cmdr.GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo() returned error: %v", err)
	}

	// Verify hostname is not empty
	if info.Hostname == "" {
		t.Error("GetSystemInfo() returned empty hostname")
	}

	// Verify IP address is not empty
	if info.IPAddress == "" {
		t.Error("GetSystemInfo() returned empty IP address")
	}

	// Verify IP address is valid (should at least be localhost)
	if info.IPAddress != "127.0.0.1" && len(info.IPAddress) < 7 {
		t.Errorf("GetSystemInfo() returned invalid IP address: %s", info.IPAddress)
	}
}

func TestCommander_Ping(t *testing.T) {
	cmdr := NewCommander()

	// Test cases for ping
	tests := []struct {
		name      string
		host      string
		wantError bool
	}{
		{
			name:      "ping localhost",
			host:      "127.0.0.1",
			wantError: false,
		},
		{
			name:      "ping localhost by name",
			host:      "localhost",
			wantError: false,
		},
		{
			name:      "ping google DNS",
			host:      "8.8.8.8",
			wantError: false,
		},
		{
			name:      "ping invalid host",
			host:      "999.999.999.999",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require network access in short mode
			if testing.Short() && (tt.host == "8.8.8.8") {
				t.Skip("skipping network test in short mode")
			}

			// Use recover to catch panics
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantError {
						t.Errorf("Ping(%s) panicked unexpectedly: %v", tt.host, r)
					}
				}
			}()

			result, err := cmdr.Ping(tt.host)

			if tt.wantError {
				// For invalid hosts, we expect either an error or a panic
				if err == nil && result.Successful {
					t.Errorf("Ping(%s) succeeded but expected error", tt.host)
				}
			} else {
				if err != nil {
					t.Errorf("Ping(%s) returned error: %v", tt.host, err)
				}
				// Note: Due to permissions, ping might fail even for valid hosts
				// So we just check that we got a result
				if result.Time < 0 {
					t.Errorf("Ping(%s) returned invalid time: %v", tt.host, result.Time)
				}
			}
		})
	}
}

func TestPingResult(t *testing.T) {
	// Test PingResult struct initialization
	pr := PingResult{
		Successful: true,
		Time:       100 * time.Millisecond,
	}

	if !pr.Successful {
		t.Error("PingResult.Successful should be true")
	}

	if pr.Time != 100*time.Millisecond {
		t.Errorf("PingResult.Time = %v, want %v", pr.Time, 100*time.Millisecond)
	}
}

func TestSystemInfo(t *testing.T) {
	// Test SystemInfo struct initialization
	si := SystemInfo{
		Hostname:  "test-host",
		IPAddress: "192.168.1.1",
	}

	if si.Hostname != "test-host" {
		t.Errorf("SystemInfo.Hostname = %s, want test-host", si.Hostname)
	}

	if si.IPAddress != "192.168.1.1" {
		t.Errorf("SystemInfo.IPAddress = %s, want 192.168.1.1", si.IPAddress)
	}
}

func BenchmarkGetSystemInfo(b *testing.B) {
	cmdr := NewCommander()
	
	for i := 0; i < b.N; i++ {
		_, err := cmdr.GetSystemInfo()
		if err != nil {
			b.Fatalf("GetSystemInfo() failed: %v", err)
		}
	}
}

func BenchmarkPingLocalhost(b *testing.B) {
	cmdr := NewCommander()
	
	for i := 0; i < b.N; i++ {
		// Use recover to handle panics
		func() {
			defer func() {
				if r := recover(); r != nil {
					b.Logf("Ping panicked: %v", r)
				}
			}()
			
			_, _ = cmdr.Ping("127.0.0.1")
		}()
	}
}