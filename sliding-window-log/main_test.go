package main
import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
func TestSlidingWindowLog_Allow(t *testing.T) {
	metrics := &Metrics{}
	sl := NewSlidingWindowLog(3, time.Minute, metrics)
	for i := 0; i < 3; i++ {
		if !sl.Allow() {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}
	if sl.Allow() {
		t.Error("Expected 4th request to be rejected")
	}
	if metrics.RejectedRequests != 1 {
		t.Errorf("Expected 1 rejected request, got %d", metrics.RejectedRequests)
	}
	time.Sleep(time.Second)
	if !sl.Allow() {
		t.Error("Expected request to be allowed after window reset")
	}
}
func TestRequestHandler(t *testing.T) {
	metrics := &Metrics{}
	sl := NewSlidingWindowLog(3, time.Minute, metrics)
	handler := RequestHandler(sl)
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
	for i := 0; i < 3; i++ {
		sl.Allow()
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
	}
}
func TestMetricsHandler(t *testing.T) {
	metrics := &Metrics{}
	handler := MetricsHandler(metrics)
	metrics.TotalRequests = 10
	metrics.RejectedRequests = 2
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := "Total requests: 10\nRejected requests: 2\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}