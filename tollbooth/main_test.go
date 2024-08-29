package main
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	tollbooth "github.com/didip/tollbooth/v7"
)
func TestEndpointHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	tlbthLimiter := tollbooth.NewLimiter(1, nil)
	tlbthLimiter.SetMessageContentType("application/json")
	http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler))
	handler := http.HandlerFunc(endpointHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := `{"status":"Successful","body":"Hi! You've reached the API. How may I help you?"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
func TestRateLimiting(t *testing.T) {
	tlbthLimiter := tollbooth.NewLimiter(1, nil)
	message := Message{
		Status: "Request Failed",
		Body:   "The API is at capacity, try again later.",
	}
	jsonMessage, _ := json.Marshal(message)
	tlbthLimiter.SetMessageContentType("application/json")
	tlbthLimiter.SetMessage(string(jsonMessage))
	handler := tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if i == 0 {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("request %d returned wrong status code: got %v want %v", i+1, status, http.StatusOK)
			}
			expected := `{"status":"Successful","body":"Hi! You've reached the API. How may I help you?"}`
			if rr.Body.String() != expected {
				t.Errorf("request %d returned unexpected body: got %v want %v", i+1, rr.Body.String(), expected)
			}
		} else {
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("request %d returned wrong status code: got %v want %v", i+1, status, http.StatusTooManyRequests)
			}
			expected := `{"status":"Request Failed","body":"The API is at capacity, try again later."}`
			if rr.Body.String() != expected {
				t.Errorf("request %d returned unexpected body: got %v want %v", i+1, rr.Body.String(), expected)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}