package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mock Commander for testing
type mockCommander struct {
	pingResult PingResult
	pingError  error
	sysInfo    SystemInfo
	sysError   error
}

func (m *mockCommander) Ping(host string) (PingResult, error) {
	if m.pingError != nil {
		return PingResult{}, m.pingError
	}
	return m.pingResult, nil
}

func (m *mockCommander) GetSystemInfo() (SystemInfo, error) {
	if m.sysError != nil {
		return SystemInfo{}, m.sysError
	}
	return m.sysInfo, nil
}

func TestHandleRequests(t *testing.T) {
	// Test that handleRequests creates a proper handler
	cmdr := &mockCommander{}
	handler := handleRequests(cmdr)
	
	if handler == nil {
		t.Fatal("handleRequests returned nil handler")
	}

	// Test that the handler is a mux with /execute endpoint
	req := httptest.NewRequest("POST", "/execute", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	// Should get some response (even if error) rather than 404
	if rec.Code == http.StatusNotFound && rec.Body.String() == "404 page not found\n" {
		t.Error("handleRequests did not register /execute endpoint")
	}
}

func TestHandleCommand_Ping(t *testing.T) {
	tests := []struct {
		name           string
		request        CommandRequest
		mockResult     PingResult
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, res CommandResponse)
	}{
		{
			name: "successful ping",
			request: CommandRequest{
				Type:    "ping",
				Payload: "google.com",
			},
			mockResult: PingResult{
				Successful: true,
				Time:       100 * time.Millisecond,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, res CommandResponse) {
				if !res.Success {
					t.Error("expected success=true")
				}
				if res.Error != "" {
					t.Errorf("unexpected error: %s", res.Error)
				}
				// Check that Data contains the time duration
				if res.Data == nil {
					t.Error("expected Data to contain time duration")
				}
			},
		},
		{
			name: "failed ping",
			request: CommandRequest{
				Type:    "ping",
				Payload: "invalid.host",
			},
			mockResult: PingResult{
				Successful: false,
				Time:       0,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, res CommandResponse) {
				if res.Success {
					t.Error("expected success=false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock commander
			cmdr := &mockCommander{
				pingResult: tt.mockResult,
				pingError:  tt.mockError,
			}

			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handler := handleCommand(cmdr)
			handler(rec, req)

			// Check status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			// Parse response
			var res CommandResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, res)
			}
		})
	}
}

func TestHandleCommand_SysInfo(t *testing.T) {
	// Setup mock commander
	expectedSysInfo := SystemInfo{
		Hostname:  "test-host",
		IPAddress: "192.168.1.100",
	}
	cmdr := &mockCommander{
		sysInfo: expectedSysInfo,
	}

	// Create request
	req := CommandRequest{
		Type:    "sysinfo",
		Payload: "",
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rec := httptest.NewRecorder()

	// Call handler
	handler := handleCommand(cmdr)
	handler(rec, httpReq)

	// Check status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Parse response
	var res CommandResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check response
	if !res.Success {
		t.Error("expected success=true")
	}
	if res.Error != "" {
		t.Errorf("unexpected error: %s", res.Error)
	}
	
	// Check that Data contains SystemInfo
	dataBytes, _ := json.Marshal(res.Data)
	var sysInfo SystemInfo
	json.Unmarshal(dataBytes, &sysInfo)
	
	if sysInfo.Hostname != expectedSysInfo.Hostname {
		t.Errorf("expected hostname %s, got %s", expectedSysInfo.Hostname, sysInfo.Hostname)
	}
	if sysInfo.IPAddress != expectedSysInfo.IPAddress {
		t.Errorf("expected IP %s, got %s", expectedSysInfo.IPAddress, sysInfo.IPAddress)
	}
}

func TestHandleCommand_InvalidType(t *testing.T) {
	cmdr := &mockCommander{}

	// Create request with invalid type
	req := CommandRequest{
		Type:    "invalid",
		Payload: "",
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rec := httptest.NewRecorder()

	// Call handler
	handler := handleCommand(cmdr)
	handler(rec, httpReq)

	// Should return 500 due to panic recovery
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestHandleCommand_InvalidJSON(t *testing.T) {
	cmdr := &mockCommander{}

	// Create request with invalid JSON
	httpReq := httptest.NewRequest("POST", "/execute", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rec := httptest.NewRecorder()

	// Call handler
	handler := handleCommand(cmdr)
	handler(rec, httpReq)

	// Should return 500 due to panic recovery
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestMiddleware_InvalidPath(t *testing.T) {
	// Create a simple handler
	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test invalid path
	req := httptest.NewRequest("POST", "/invalid", nil)
	rec := httptest.NewRecorder()
	
	handler(rec, req)

	// Should return 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestMiddleware_InvalidMethod(t *testing.T) {
	// Create a simple handler
	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test invalid method
	req := httptest.NewRequest("GET", "/execute", nil)
	rec := httptest.NewRecorder()
	
	handler(rec, req)

	// Should return 405
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestMiddleware_PanicRecovery(t *testing.T) {
	// Create a handler that panics
	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Test panic recovery
	req := httptest.NewRequest("POST", "/execute", nil)
	rec := httptest.NewRecorder()
	
	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Error("middleware did not recover from panic")
		}
	}()
	
	handler(rec, req)

	// Should return 500
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestCommandRequest(t *testing.T) {
	// Test CommandRequest JSON marshaling/unmarshaling
	req := CommandRequest{
		Type:    "ping",
		Payload: "example.com",
	}

	// Marshal
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal CommandRequest: %v", err)
	}

	// Unmarshal
	var decoded CommandRequest
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal CommandRequest: %v", err)
	}

	// Verify
	if decoded.Type != req.Type {
		t.Errorf("Type mismatch: expected %s, got %s", req.Type, decoded.Type)
	}
	if decoded.Payload != req.Payload {
		t.Errorf("Payload mismatch: expected %s, got %s", req.Payload, decoded.Payload)
	}
}

func TestCommandResponse(t *testing.T) {
	// Test CommandResponse JSON marshaling/unmarshaling
	res := CommandResponse{
		Success: true,
		Data:    "test data",
		Error:   "",
	}

	// Marshal
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal CommandResponse: %v", err)
	}

	// Unmarshal
	var decoded CommandResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal CommandResponse: %v", err)
	}

	// Verify
	if decoded.Success != res.Success {
		t.Errorf("Success mismatch: expected %v, got %v", res.Success, decoded.Success)
	}
	if decoded.Error != res.Error {
		t.Errorf("Error mismatch: expected %s, got %s", res.Error, decoded.Error)
	}

	// Test with error field (should be omitted when empty)
	if bytes.Contains(data, []byte("error")) {
		t.Error("error field should be omitted when empty")
	}

	// Test with non-empty error
	res.Error = "test error"
	data, _ = json.Marshal(res)
	if !bytes.Contains(data, []byte("error")) {
		t.Error("error field should be present when non-empty")
	}
}

func BenchmarkHandleCommand_Ping(b *testing.B) {
	cmdr := &mockCommander{
		pingResult: PingResult{
			Successful: true,
			Time:       100 * time.Millisecond,
		},
	}

	req := CommandRequest{
		Type:    "ping",
		Payload: "test.com",
	}
	body, _ := json.Marshal(req)
	
	handler := handleCommand(cmdr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler(rec, httpReq)
	}
}

func BenchmarkHandleCommand_SysInfo(b *testing.B) {
	cmdr := &mockCommander{
		sysInfo: SystemInfo{
			Hostname:  "test-host",
			IPAddress: "192.168.1.1",
		},
	}

	req := CommandRequest{
		Type:    "sysinfo",
		Payload: "",
	}
	body, _ := json.Marshal(req)
	
	handler := handleCommand(cmdr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest("POST", "/execute", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler(rec, httpReq)
	}
}