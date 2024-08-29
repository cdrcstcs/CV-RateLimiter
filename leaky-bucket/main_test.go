package main
import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
func TestLeakyBucket_Allow(t *testing.T) {
	metrics := &Metrics{}
	bucket := NewLeakyBucket(2, 100*time.Millisecond, metrics)
	if !bucket.Allow() {
		t.Fatal("expected to allow the first request")
	}
	if !bucket.Allow() {
		t.Fatal("expected to allow the second request")
	}
	if bucket.Allow() {
		t.Fatal("expected to reject the third request")
	}
	time.Sleep(150 * time.Millisecond)
	if !bucket.Allow() {
		t.Fatal("expected to allow a request after the bucket has leaked")
	}
}
func TestRequestHandler(t *testing.T) {
	metrics := &Metrics{}
	bucket := NewLeakyBucket(2, 100*time.Millisecond, metrics)
	handler := RequestHandler(bucket)
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
	bucket := NewLeakyBucket(2, 100*time.Millisecond, metrics)
	handler := MetricsHandler(metrics)
	bucket.Allow()
	bucket.Allow()
	bucket.Allow()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	expected := "Total processed requests: 2\nTotal discarded requests: 1\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
func TestLeakyBucketLeak(t *testing.T) {
	metrics := &Metrics{}
	bucket := NewLeakyBucket(1, 100*time.Millisecond, metrics)
	if !bucket.Allow() {
		t.Fatal("expected to allow the first request")
	}
	time.Sleep(200 * time.Millisecond)
	if !bucket.Allow() {
		t.Fatal("expected to allow a request after the bucket has leaked")
	}
}