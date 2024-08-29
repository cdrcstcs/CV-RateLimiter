package main
import (
	"fmt"
	"net/http"
	"sync"
	"time"
)
type FixedWindowCounter struct {
	limit          int          
	windowDuration time.Duration
	count          int           
	resetTime      time.Time    
	mutex          sync.Mutex  
	metrics        *Metrics
}
type Metrics struct {
	TotalRequests  int
	RejectedRequests int 
	Mutex          sync.Mutex
}
func NewFixedWindowCounter(limit int, windowDuration time.Duration, metrics *Metrics) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:          limit,
		windowDuration: windowDuration,
		count:          0,
		resetTime:      time.Now().Add(windowDuration),
		metrics:        metrics,
	}
}
func (fw *FixedWindowCounter) Allow() bool {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()
	now := time.Now()
	if now.After(fw.resetTime) {
		fw.count = 0
		fw.resetTime = now.Add(fw.windowDuration)
	}
	if fw.count < fw.limit {
		fw.count++
		fw.metrics.Mutex.Lock()
		fw.metrics.TotalRequests++
		fw.metrics.Mutex.Unlock()
		return true
	}
	fw.metrics.Mutex.Lock()
	fw.metrics.RejectedRequests++
	fw.metrics.Mutex.Unlock()
	return false
}
func RequestHandler(counter *FixedWindowCounter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if counter.Allow() {
			fmt.Fprintf(w, "Request processed\n")
		} else {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		}
	}
}
func MetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.Mutex.Lock()
		defer metrics.Mutex.Unlock()
		fmt.Fprintf(w, "Total requests: %d\n", metrics.TotalRequests)
		fmt.Fprintf(w, "Rejected requests: %d\n", metrics.RejectedRequests)
	}
}
func main() {
	metrics := &Metrics{}
	counter := NewFixedWindowCounter(100, time.Minute, metrics)
	http.HandleFunc("/", RequestHandler(counter))
	http.HandleFunc("/metrics", MetricsHandler(metrics))
	server := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Server is running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server failed:", err)
	}
}