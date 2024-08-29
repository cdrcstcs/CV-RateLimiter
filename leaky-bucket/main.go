package main
import (
	"fmt"
	"net/http"
	"sync"
	"time"
)
type LeakyBucket struct {
	capacity       int          
	water          int       
	leakRate       time.Duration
	lastLeakTime   time.Time    
	mutex          sync.Mutex   
	metrics        *Metrics 
}
type Metrics struct {
	ProcessedCount int 
	DiscardedCount int 
	Mutex          sync.Mutex
}
func NewLeakyBucket(capacity int, leakRate time.Duration, metrics *Metrics) *LeakyBucket {
	return &LeakyBucket{
		capacity:     capacity,
		water:        0,
		leakRate:     leakRate,
		lastLeakTime: time.Now(),
		metrics:      metrics,
	}
}
func (b *LeakyBucket) leak() {
	now := time.Now()
	elapsed := now.Sub(b.lastLeakTime)
	leaked := int(elapsed / b.leakRate)
	if leaked > 0 {
		b.water -= leaked
		if b.water < 0 {
			b.water = 0
		}
		b.lastLeakTime = now
	}
}
func (b *LeakyBucket) Allow() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.leak()
	if b.water < b.capacity {
		b.water++
		b.metrics.Mutex.Lock()
		b.metrics.ProcessedCount++
		b.metrics.Mutex.Unlock()
		return true
	}
	b.metrics.Mutex.Lock()
	b.metrics.DiscardedCount++
	b.metrics.Mutex.Unlock()
	return false
}
func RequestHandler(bucket *LeakyBucket) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if bucket.Allow() {
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
		fmt.Fprintf(w, "Total processed requests: %d\n", metrics.ProcessedCount)
		fmt.Fprintf(w, "Total discarded requests: %d\n", metrics.DiscardedCount)
	}
}
func main() {
	metrics := &Metrics{}
	bucket := NewLeakyBucket(10, time.Second, metrics)
	http.HandleFunc("/", RequestHandler(bucket))
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