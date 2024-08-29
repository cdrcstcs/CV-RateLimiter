package main
import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
func TestSlidingWindowCounter_Allow(t *testing.T) {
	metrics := &Metrics{}
	counter := NewSlidingWindowCounter(2, time.Minute, 2, time.Second, metrics)
	if !counter.Allow() {
		t.Error("Expected request to be allowed, but it was not")
	}
	if !counter.Allow() {
		t.Error("Expected request to be allowed, but it was not")
	}
	if counter.Allow() {
		t.Error("Expected request to be rejected, but it was allowed")
	}
	time.Sleep(2 * time.Second)
	if !counter.Allow() {
		t.Error("Expected request to be allowed after waiting, but it was not")
	}
}
func TestRequestHandler(t *testing.T) {
	metrics := &Metrics{}
	counter := NewSlidingWindowCounter(2, time.Minute, 2, time.Second, metrics)
	handler := RequestHandler(counter)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if status := rec.Result().StatusCode; status != http.StatusOK {
		t.Errorf("Expected status OK but got %v", status)
	}
	expectedBody := "Request processed\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q but got %q", expectedBody, rec.Body.String())
	}
}
func TestMetricsHandler(t *testing.T) {
	metrics := &Metrics{}
	counter := NewSlidingWindowCounter(2, time.Minute, 2, time.Second, metrics)
	handler := MetricsHandler(metrics)
	for i := 0; i < 3; i++ {
		counter.Allow()
	}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if status := rec.Result().StatusCode; status != http.StatusOK {
		t.Errorf("Expected status OK but got %v", status)
	}
	expectedBody := "Total requests: 3\nRejected requests: 1\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q but got %q", expectedBody, rec.Body.String())
	}
}