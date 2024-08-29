package main
import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
type MockTransport struct{}
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, net.ErrClosed 
}
func TestRateLimit(t *testing.T) {
	handler := perClientRateLimiter(endpointHandler)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	for i := 0; i < 5; i++ {
		handler.ServeHTTP(rr, req)
		if i < 2 {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("expected status OK, got %v", status)
			}
		} else {
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("expected status Too Many Requests, got %v", status)
			}
			var message Message
			if err := json.NewDecoder(rr.Body).Decode(&message); err != nil {
				t.Fatalf("could not decode response body: %v", err)
			}
			if message.Status != "Request Failed" {
				t.Errorf("expected status 'Request Failed', got %v", message.Status)
			}
		}
		rr = httptest.NewRecorder()
	}
}
func TestClientRateLimiter_Clearing(t *testing.T) {
	handler := perClientRateLimiter(endpointHandler)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	time.Sleep(2 * time.Minute)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}
	time.Sleep(2 * time.Minute)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK after cleanup, got %v", status)
	}
}
func TestEndpointHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	endpointHandler(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}
	var message Message
	if err := json.NewDecoder(rr.Body).Decode(&message); err != nil {
		t.Fatalf("could not decode response body: %v", err)
	}
	if message.Status != "Successful" {
		t.Errorf("expected status 'Successful', got %v", message.Status)
	}
	if message.Body != "Hi! You've reached the API. How may I help you?" {
		t.Errorf("expected body 'Hi! You've reached the API. How may I help you?', got %v", message.Body)
	}
}
func TestPerClientRateLimiter_InvalidRemoteAddr(t *testing.T) {
	originalTransport := http.DefaultTransport
	http.DefaultTransport = &MockTransport{}
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	handler := perClientRateLimiter(endpointHandler)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected status Internal Server Error, got %v", status)
	}
}