package main
import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
func TestTokenBucket_Allow(t *testing.T) {
	metrics := &Metrics{}
	bucket := NewTokenBucket(3, time.Second, metrics)
	for i := 0; i < 3; i++ {
		if !bucket.Allow() {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}
	if bucket.Allow() {
		t.Error("Expected 4th request to be rejected")
	}
	if metrics.RejectedCount != 1 {
		t.Errorf("Expected 1 rejected request, got %d", metrics.RejectedCount)
	}
	time.Sleep(2 * time.Second)
	if !bucket.Allow() {
		t.Error("Expected request to be allowed after refill")
	}
}
func TestRequestHandler(t *testing.T) {
	metrics := &Metrics{}
	bucket := NewTokenBucket(3, time.Second, metrics)
	handler := RequestHandler(bucket)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := "Request allowed\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
	for i := 0; i < 3; i++ {
		bucket.Allow()
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
	metrics.RequestCount = 10
	metrics.RejectedCount = 2
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
func TestAdminBucket(t *testing.T) {
	metrics := &Metrics{}
	bucket := NewTokenBucket(5, 500*time.Millisecond, metrics)
	handler := RequestHandler(bucket)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := "Request allowed\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
	for i := 0; i < 5; i++ {
		bucket.Allow()
	}
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
	}
}