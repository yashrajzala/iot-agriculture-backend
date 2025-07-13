package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"iot-agriculture-backend/internal/config"
	"iot-agriculture-backend/internal/models"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// CircuitBreaker states
const (
	StateClosed = iota
	StateOpen
	StateHalfOpen
)

// InfluxDBService handles InfluxDB operations
type InfluxDBService struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	org      string
	bucket   string
	config   *config.InfluxDBConfig

	// Circuit breaker
	mu              sync.RWMutex
	state           int
	failureCount    int
	lastFailureTime time.Time
	threshold       int
	timeout         time.Duration
}

// NewInfluxDBService creates a new InfluxDB service
func NewInfluxDBService(cfg *config.InfluxDBConfig) *InfluxDBService {
	// Validate required configuration
	if cfg.Token == "" {
		log.Printf("Warning: INFLUXDB_TOKEN not set - InfluxDB logging will be disabled")
		return &InfluxDBService{
			client:   nil,
			writeAPI: nil,
			org:      cfg.Org,
			bucket:   cfg.Bucket,
			config:   cfg,
		}
	}

	// Create client with optimized settings
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	defer client.Close()

	// Create write API
	writeAPI := client.WriteAPIBlocking(cfg.Org, cfg.Bucket)

	// Test connection
	_, err := client.Ping(context.Background())
	if err != nil {
		log.Printf("Warning: Could not connect to InfluxDB: %v", err)
		log.Printf("InfluxDB logging will be disabled")
		return &InfluxDBService{
			client:   nil,
			writeAPI: nil,
			org:      cfg.Org,
			bucket:   cfg.Bucket,
			config:   cfg,
		}
	}

	log.Printf("Successfully connected to InfluxDB at %s", cfg.URL)
	log.Printf("Using organization: %s, bucket: %s", cfg.Org, cfg.Bucket)
	return &InfluxDBService{
		client:    client,
		writeAPI:  writeAPI,
		org:       cfg.Org,
		bucket:    cfg.Bucket,
		config:    cfg,
		state:     StateClosed,
		threshold: 5,                // Fail after 5 consecutive failures
		timeout:   30 * time.Second, // Wait 30 seconds before trying again
	}
}

// LogAverages logs sensor averages to InfluxDB with circuit breaker
func (i *InfluxDBService) LogAverages(averages models.AverageResult) error {
	if i.client == nil || i.writeAPI == nil {
		return fmt.Errorf("InfluxDB not connected")
	}

	// Check circuit breaker state
	if !i.canExecute() {
		return fmt.Errorf("circuit breaker is open - InfluxDB writes are temporarily disabled")
	}

	// Create point for sensor averages
	point := influxdb2.NewPoint(
		"sensor_averages",
		map[string]string{
			"greenhouse_id": averages.GreenhouseID,
			"node_id":       averages.NodeID,
		},
		map[string]interface{}{
			"s1_average": averages.S1Average,
			"s2_average": averages.S2Average,
			"s3_average": averages.S3Average,
			"s4_average": averages.S4Average,
			"s5_average": averages.S5Average,
			"s6_average": averages.S6Average,
			"s7_average": averages.S7Average,
			"s8_average": averages.S8Average,
			"s9_average": averages.S9Average,
			"readings":   averages.Readings,
			"duration":   averages.Duration,
		},
		time.Now(),
	)

	// Write point to InfluxDB
	err := i.writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		i.recordFailure()
		return fmt.Errorf("failed to write to InfluxDB: %w", err)
	}

	i.recordSuccess()
	log.Printf("Logged sensor averages to InfluxDB: %s/%s (%.1fs, %d readings)",
		averages.GreenhouseID, averages.NodeID, averages.Duration, averages.Readings)
	return nil
}

// canExecute checks if the circuit breaker allows execution
func (i *InfluxDBService) canExecute() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	switch i.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(i.lastFailureTime) > i.timeout {
			i.mu.RUnlock()
			i.mu.Lock()
			i.state = StateHalfOpen
			i.mu.Unlock()
			i.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure records a failure and updates circuit breaker state
func (i *InfluxDBService) recordFailure() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.failureCount++
	i.lastFailureTime = time.Now()

	if i.state == StateClosed && i.failureCount >= i.threshold {
		i.state = StateOpen
		log.Printf("Circuit breaker opened - InfluxDB writes disabled for %v", i.timeout)
	} else if i.state == StateHalfOpen {
		i.state = StateOpen
		log.Printf("Circuit breaker reopened - InfluxDB writes disabled for %v", i.timeout)
	}
}

// recordSuccess records a success and resets circuit breaker
func (i *InfluxDBService) recordSuccess() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.state == StateHalfOpen {
		i.state = StateClosed
		i.failureCount = 0
		log.Printf("Circuit breaker closed - InfluxDB writes re-enabled")
	}
}

// Note: Individual sensor logging removed - only averages are logged every 60 seconds

// Close closes the InfluxDB connection
func (i *InfluxDBService) Close() {
	if i.client != nil {
		i.client.Close()
		log.Println("InfluxDB connection closed")
	}
}

// IsConnected returns true if InfluxDB is connected
func (i *InfluxDBService) IsConnected() bool {
	connected := i.client != nil && i.writeAPI != nil
	// Update metrics if available (will be set by sensor service)
	return connected
}

// GetConnectionInfo returns connection information
func (i *InfluxDBService) GetConnectionInfo() string {
	if i.IsConnected() {
		return fmt.Sprintf("Connected to InfluxDB - Org: %s, Bucket: %s", i.org, i.bucket)
	}
	return "InfluxDB not connected"
}
