package main
import (
	"fmt"
	"net/http"
	"sync"
	"time"
)
type SlidingWindowCounter struct {
	limit          int           
	windowDuration time.Duration 
	windowSize     int           
	interval       time.Duration 
	counters       []int         
	currentIndex   int           
	mutex          sync.Mutex    
	metrics        *Metrics      
}
type Metrics struct {
	TotalRequests  int 
	RejectedRequests int 
	Mutex          sync.Mutex
}
func NewSlidingWindowCounter(limit int, windowDuration time.Duration, windowSize int, interval time.Duration, metrics *Metrics) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:          limit,
		windowDuration: windowDuration,
		windowSize:     windowSize,
		interval:       interval,
		counters:       make([]int, windowSize),
		currentIndex:   0,
		metrics:        metrics,
	}
}
func (s *SlidingWindowCounter) Allow() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	now := time.Now()
	windowStart := now.Add(-s.windowDuration)
	for i := 0; i < s.windowSize; i++ {
		intervalStart := now.Add(-time.Duration(i) * s.interval)
		if intervalStart.Before(windowStart) {
			s.counters[i] = 0
		}
	}
	if s.counters[s.currentIndex] < s.limit {
		s.counters[s.currentIndex]++
		s.metrics.Mutex.Lock()
		s.metrics.TotalRequests++
		s.metrics.Mutex.Unlock()
		return true
	}
	s.currentIndex = (s.currentIndex + 1) % s.windowSize
	s.metrics.Mutex.Lock()
	s.metrics.RejectedRequests++
	s.metrics.Mutex.Unlock()
	return false
}
func RequestHandler(counter *SlidingWindowCounter) http.HandlerFunc {
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
	counter := NewSlidingWindowCounter(100, time.Minute, 60, time.Second, metrics)
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