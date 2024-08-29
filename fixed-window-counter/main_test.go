package main
import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
func TestFixedWindowCounter_Allow(t *testing.T) {
	metrics := &Metrics{}
	counter := NewFixedWindowCounter(2, time.Second, metrics)
	if !counter.Allow() {
		t.Fatal("expected to allow the first request")
	}
	if !counter.Allow() {
		t.Fatal("expected to allow the second request")
	}
	if counter.Allow() {
		t.Fatal("expected to reject the third request")
	}
	time.Sleep(1 * time.Second)
	if !counter.Allow() {
		t.Fatal("expected to allow a request after window reset")
	}
}
func TestRequestHandler(t *testing.T) {
	metrics := &Metrics{}
	counter := NewFixedWindowCounter(2, time.Second, metrics)
	handler := RequestHandler(counter)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := "Request processed\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
func TestMetricsHandler(t *testing.T) {
	metrics := &Metrics{}
	counter := NewFixedWindowCounter(2, time.Second, metrics)
	handler := MetricsHandler(metrics)
	counter.Allow()
	counter.Allow()
	counter.Allow() 
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	expected := "Total requests: 3\nRejected requests: 1\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
func TestRateLimitReset(t *testing.T) {
	metrics := &Metrics{}
	counter := NewFixedWindowCounter(1, time.Second, metrics)
	if !counter.Allow() {
		t.Fatal("expected to allow the first request")
	}
	time.Sleep(2 * time.Second)
	if !counter.Allow() {
		t.Fatal("expected to allow a request after the rate limit window has reset")
	}
}