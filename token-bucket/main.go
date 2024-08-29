package main
import (
	"fmt"
	"net/http"
	"sync"
	"time"
)
type TokenBucket struct {
	capacity     int
	tokens       int
	rate         time.Duration
	lastRefill   time.Time
	mutex        sync.Mutex
	metrics      *Metrics
}
type Metrics struct {
	RequestCount    int
	RejectedCount   int
	Mutex           sync.Mutex
}
func NewTokenBucket(capacity int, rate time.Duration, metrics *Metrics) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastRefill: time.Now(),
		metrics:    metrics,
	}
}
func (b *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	newTokens := int(elapsed / b.rate)
	if newTokens > 0 {
		b.tokens += newTokens
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefill = now
	}
}
func (b *TokenBucket) Allow() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.refill()
	if b.tokens > 0 {
		b.tokens--
		b.metrics.Mutex.Lock()
		b.metrics.RequestCount++
		b.metrics.Mutex.Unlock()
		return true
	}
	b.metrics.Mutex.Lock()
	b.metrics.RejectedCount++
	b.metrics.Mutex.Unlock()
	return false
}
func RequestHandler(bucket *TokenBucket) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if bucket.Allow() {
			fmt.Fprintf(w, "Request allowed\n")
		} else {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		}
	}
}
func MetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.Mutex.Lock()
		defer metrics.Mutex.Unlock()
		fmt.Fprintf(w, "Total requests: %d\n", metrics.RequestCount)
		fmt.Fprintf(w, "Rejected requests: %d\n", metrics.RejectedCount)
	}
}
func main() {
	metrics := &Metrics{}
	globalBucket := NewTokenBucket(10, time.Second, metrics)
	adminBucket := NewTokenBucket(5, 500*time.Millisecond, metrics)
	http.HandleFunc("/", RequestHandler(globalBucket))
	http.HandleFunc("/admin", RequestHandler(adminBucket))
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