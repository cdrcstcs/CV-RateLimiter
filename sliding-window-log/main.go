package main
import (
	"container/list"
	"fmt"
	"net/http"
	"sync"
	"time"
)
type SlidingWindowLog struct {
	limit          int           
	windowDuration time.Duration 
	requests       *list.List    
	mutex          sync.Mutex    
	metrics        *Metrics     
}
type Metrics struct {
	TotalRequests  int 
	RejectedRequests int 
	Mutex          sync.Mutex
}
func NewSlidingWindowLog(limit int, windowDuration time.Duration, metrics *Metrics) *SlidingWindowLog {
	return &SlidingWindowLog{
		limit:          limit,
		windowDuration: windowDuration,
		requests:       list.New(),
		metrics:        metrics,
	}
}
func (s *SlidingWindowLog) Allow() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	now := time.Now()
	windowStart := now.Add(-s.windowDuration)
	for s.requests.Len() > 0 {
		oldest := s.requests.Front()
		if oldest.Value.(time.Time).Before(windowStart) {
			s.requests.Remove(oldest)
		} else {
			break
		}
	}
	if s.requests.Len() < s.limit {
		s.requests.PushBack(now)
		s.metrics.Mutex.Lock()
		s.metrics.TotalRequests++
		s.metrics.Mutex.Unlock()
		return true
	}
	s.metrics.Mutex.Lock()
	s.metrics.RejectedRequests++
	s.metrics.Mutex.Unlock()
	return false
}
func RequestHandler(sl *SlidingWindowLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sl.Allow() {
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
	sl := NewSlidingWindowLog(100, time.Minute, metrics)
	http.HandleFunc("/", RequestHandler(sl))
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