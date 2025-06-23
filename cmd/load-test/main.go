package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// TestResult holds the results of a load test
type TestResult struct {
	Mode              string        `json:"mode"`
	ConcurrentUsers   int           `json:"concurrent_users"`
	Duration          time.Duration `json:"duration"`
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulReqs    int64         `json:"successful_requests"`
	FailedReqs        int64         `json:"failed_requests"`
	AvgResponseTime   time.Duration `json:"avg_response_time"`
	MinResponseTime   time.Duration `json:"min_response_time"`
	MaxResponseTime   time.Duration `json:"max_response_time"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	Throughput        float64       `json:"throughput_mbps"`
	ErrorRate         float64       `json:"error_rate"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
}

// LoadTester manages the load testing process
type LoadTester struct {
	baseURL         string
	concurrentUsers int
	duration        time.Duration
	rampUp          time.Duration
	mode            string
	client          *http.Client
	results         *TestResult
	mu              sync.Mutex
	responseTimes   []time.Duration
}

// NewLoadTester creates a new load tester instance
func NewLoadTester(baseURL, mode string, users int, duration, rampUp time.Duration) *LoadTester {
	return &LoadTester{
		baseURL:         baseURL,
		concurrentUsers: users,
		duration:        duration,
		rampUp:          rampUp,
		mode:            mode,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		results: &TestResult{
			Mode:            mode,
			ConcurrentUsers: users,
			Duration:        duration,
			MinResponseTime: time.Hour, // Will be updated with actual min
		},
		responseTimes: make([]time.Duration, 0),
	}
}

// Run executes the load test
func (lt *LoadTester) Run(ctx context.Context) (*TestResult, error) {
	log.Printf("Starting load test: mode=%s, users=%d, duration=%v", lt.mode, lt.concurrentUsers, lt.duration)

	lt.results.StartTime = time.Now()
	defer func() {
		lt.results.EndTime = time.Now()
		lt.calculateFinalMetrics()
	}()

	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, lt.duration)
	defer cancel()

	// Counters for metrics
	var totalRequests, successfulReqs, failedReqs int64

	// Channel to control user ramp-up
	userChan := make(chan struct{}, lt.concurrentUsers)
	rampUpInterval := lt.rampUp / time.Duration(lt.concurrentUsers)

	// WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Start users with ramp-up
	go func() {
		for i := 0; i < lt.concurrentUsers; i++ {
			select {
			case <-testCtx.Done():
				return
			case userChan <- struct{}{}:
				wg.Add(1)
				go lt.runUser(testCtx, &wg, &totalRequests, &successfulReqs, &failedReqs)
				time.Sleep(rampUpInterval)
			}
		}
	}()

	// Wait for test completion
	wg.Wait()

	// Update final counters
	lt.results.TotalRequests = atomic.LoadInt64(&totalRequests)
	lt.results.SuccessfulReqs = atomic.LoadInt64(&successfulReqs)
	lt.results.FailedReqs = atomic.LoadInt64(&failedReqs)

	return lt.results, nil
}

// runUser simulates a single user making requests
func (lt *LoadTester) runUser(ctx context.Context, wg *sync.WaitGroup, totalReqs, successReqs, failReqs *int64) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			lt.makeRequest(totalReqs, successReqs, failReqs)
			time.Sleep(100 * time.Millisecond) // Small delay between requests
		}
	}
}

// makeRequest makes a single HTTP request and records metrics
func (lt *LoadTester) makeRequest(totalReqs, successReqs, failReqs *int64) {
	atomic.AddInt64(totalReqs, 1)

	// Choose endpoint based on mode
	var endpoint string
	switch lt.mode {
	case "actor":
		endpoint = "/api/v1/rides/request"
	case "traditional":
		endpoint = "/api/v1/traditional/rides/request"
	default:
		endpoint = "/health"
	}

	url := lt.baseURL + endpoint

	// Create request payload
	payload := map[string]interface{}{
		"passenger_id": fmt.Sprintf("passenger_%d", time.Now().UnixNano()%1000),
		"pickup_location": map[string]float64{
			"latitude":  -6.2088 + (float64(time.Now().UnixNano()%100) / 10000),
			"longitude": 106.8456 + (float64(time.Now().UnixNano()%100) / 10000),
		},
		"destination_location": map[string]float64{
			"latitude":  -6.1944 + (float64(time.Now().UnixNano()%100) / 10000),
			"longitude": 106.8229 + (float64(time.Now().UnixNano()%100) / 10000),
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	// Make request and measure time
	start := time.Now()
	resp, err := http.Post(url, "application/json", nil)
	responseTime := time.Since(start)

	// Record response time
	lt.mu.Lock()
	lt.responseTimes = append(lt.responseTimes, responseTime)
	if responseTime < lt.results.MinResponseTime {
		lt.results.MinResponseTime = responseTime
	}
	if responseTime > lt.results.MaxResponseTime {
		lt.results.MaxResponseTime = responseTime
	}
	lt.mu.Unlock()

	if err != nil {
		atomic.AddInt64(failReqs, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(successReqs, 1)
	} else {
		atomic.AddInt64(failReqs, 1)
	}

	_ = payloadBytes // Use the payload bytes (placeholder)
}

// calculateFinalMetrics calculates final performance metrics
func (lt *LoadTester) calculateFinalMetrics() {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if len(lt.responseTimes) == 0 {
		return
	}

	// Calculate average response time
	var totalTime time.Duration
	for _, rt := range lt.responseTimes {
		totalTime += rt
	}
	lt.results.AvgResponseTime = totalTime / time.Duration(len(lt.responseTimes))

	// Calculate requests per second
	actualDuration := lt.results.EndTime.Sub(lt.results.StartTime)
	lt.results.RequestsPerSecond = float64(lt.results.TotalRequests) / actualDuration.Seconds()

	// Calculate error rate
	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedReqs) / float64(lt.results.TotalRequests) * 100
	}

	// Estimate throughput (simplified calculation)
	avgPayloadSize := 200.0 // bytes (estimated)
	lt.results.Throughput = (float64(lt.results.SuccessfulReqs) * avgPayloadSize * 8) / (1024 * 1024) / actualDuration.Seconds()
}

// PrintResults prints the test results in a formatted way
func (lt *LoadTester) PrintResults() {
	fmt.Printf("\n=== Load Test Results ===\n")
	fmt.Printf("Mode: %s\n", lt.results.Mode)
	fmt.Printf("Concurrent Users: %d\n", lt.results.ConcurrentUsers)
	fmt.Printf("Test Duration: %v\n", lt.results.Duration)
	fmt.Printf("Total Requests: %d\n", lt.results.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", lt.results.SuccessfulReqs)
	fmt.Printf("Failed Requests: %d\n", lt.results.FailedReqs)
	fmt.Printf("Requests/Second: %.2f\n", lt.results.RequestsPerSecond)
	fmt.Printf("Average Response Time: %v\n", lt.results.AvgResponseTime)
	fmt.Printf("Min Response Time: %v\n", lt.results.MinResponseTime)
	fmt.Printf("Max Response Time: %v\n", lt.results.MaxResponseTime)
	fmt.Printf("Error Rate: %.2f%%\n", lt.results.ErrorRate)
	fmt.Printf("Throughput: %.2f Mbps\n", lt.results.Throughput)
	fmt.Printf("========================\n\n")
}

// SaveResults saves the test results to a JSON file
func (lt *LoadTester) SaveResults(filename string) error {
	data, err := json.MarshalIndent(lt.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func main() {
	// Command line flags
	var (
		mode     = flag.String("mode", "actor", "Test mode: actor or traditional")
		users    = flag.Int("users", 10, "Number of concurrent users")
		duration = flag.Duration("duration", 1*time.Minute, "Test duration")
		rampUp   = flag.Duration("rampup", 10*time.Second, "Ramp-up duration")
		baseURL  = flag.String("url", "http://localhost:8080", "Base URL for testing")
		output   = flag.String("output", "", "Output file for results (JSON)")
		config   = flag.String("config", "", "Config file path")
	)
	flag.Parse()

	// Load configuration if provided (simplified for now)
	if *config != "" {
		log.Printf("Config file specified: %s (config loading not implemented yet)", *config)
	}

	// Validate mode
	if *mode != "actor" && *mode != "traditional" {
		log.Fatal("Mode must be either 'actor' or 'traditional'")
	}

	// Create load tester
	tester := NewLoadTester(*baseURL, *mode, *users, *duration, *rampUp)

	// Run the test
	ctx := context.Background()
	_, err := tester.Run(ctx)
	if err != nil {
		log.Fatalf("Load test failed: %v", err)
	}

	// Print results
	tester.PrintResults()

	// Save results if output file specified
	if *output != "" {
		if err := tester.SaveResults(*output); err != nil {
			log.Printf("Failed to save results: %v", err)
		} else {
			log.Printf("Results saved to: %s", *output)
		}
	}

	// Auto-save with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("load-test-results/load_test_%s_%s.json", *mode, timestamp)
	os.MkdirAll("load-test-results", 0755)
	if err := tester.SaveResults(filename); err != nil {
		log.Printf("Failed to auto-save results: %v", err)
	} else {
		log.Printf("Results auto-saved to: %s", filename)
	}

	log.Println("Load test completed successfully")
}