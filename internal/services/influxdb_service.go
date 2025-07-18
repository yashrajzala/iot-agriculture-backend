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
	writeAPI api.WriteAPIBlocking // Reverted to blocking API for reliability
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

	// Shutdown protection
	shutdownMu sync.RWMutex
	shutdown   bool
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

	// Create blocking write API for reliability
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
	log.Printf("Blocking writes enabled for reliability")
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
	// Check shutdown state first
	i.shutdownMu.RLock()
	if i.shutdown {
		i.shutdownMu.RUnlock()
		return fmt.Errorf("InfluxDB service is shutting down")
	}
	i.shutdownMu.RUnlock()

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

	// Write point to InfluxDB with blocking API
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
	// Mark as shutting down first
	i.shutdownMu.Lock()
	i.shutdown = true
	i.shutdownMu.Unlock()

	// Close the client
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

// GetLatestAveragesFromDB fetches the latest average for each node from InfluxDB
func (i *InfluxDBService) GetLatestAveragesFromDB(greenhouseID, nodeID string) ([]models.AverageResult, error) {
	if i.client == nil || i.writeAPI == nil {
		return nil, fmt.Errorf("InfluxDB not connected")
	}
	q := `from(bucket: "` + i.bucket + `")
	  |> range(start: -7d)
	  |> filter(fn: (r) => r._measurement == "sensor_averages")`
	if greenhouseID != "" {
		q += ` |> filter(fn: (r) => r.greenhouse_id == "` + greenhouseID + `")`
	}
	if nodeID != "" {
		q += ` |> filter(fn: (r) => r.node_id == "` + nodeID + `")`
	}
	q += ` |> sort(columns: ["_time"], desc: true)`
	q += ` |> group(columns: ["greenhouse_id", "node_id", "_field"])
	  |> first()`

	queryAPI := i.client.QueryAPI(i.org)
	result, err := queryAPI.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}

	type nodeKey struct{ GreenhouseID, NodeID string }
	// Map: nodeKey -> field -> value
	nodeMap := make(map[nodeKey]map[string]float64)
	nodeTime := make(map[nodeKey]time.Time)
	for result.Next() {
		gID := result.Record().ValueByKey("greenhouse_id")
		nID := result.Record().ValueByKey("node_id")
		field := result.Record().Field()
		value, ok := result.Record().Value().(float64)
		if !ok {
			continue
		}
		key := nodeKey{fmt.Sprint(gID), fmt.Sprint(nID)}
		if _, ok := nodeMap[key]; !ok {
			nodeMap[key] = make(map[string]float64)
		}
		nodeMap[key][field] = value
		t := result.Record().Time()
		if t.After(nodeTime[key]) {
			nodeTime[key] = t
		}
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var out []models.AverageResult
	for key, fields := range nodeMap {
		out = append(out, models.AverageResult{
			GreenhouseID: key.GreenhouseID,
			NodeID:       key.NodeID,
			S1Average:    fields["s1_average"],
			S2Average:    fields["s2_average"],
			S3Average:    fields["s3_average"],
			S4Average:    fields["s4_average"],
			S5Average:    fields["s5_average"],
			S6Average:    fields["s6_average"],
			S7Average:    fields["s7_average"],
			S8Average:    fields["s8_average"],
			S9Average:    fields["s9_average"],
			// Duration, Readings, etc. can be added if stored in DB
			// Timestamp: nodeTime[key],
		})
	}
	return out, nil
}

// GetAllAveragesFromDB fetches all average data for all nodes from InfluxDB
func (i *InfluxDBService) GetAllAveragesFromDB(greenhouseID, nodeID string) ([]models.AverageResult, error) {
	if i.client == nil || i.writeAPI == nil {
		return nil, fmt.Errorf("InfluxDB not connected")
	}
	q := `from(bucket: "` + i.bucket + `")
	  |> range(start: -30d)
	  |> filter(fn: (r) => r._measurement == "sensor_averages")`
	if greenhouseID != "" {
		q += ` |> filter(fn: (r) => r.greenhouse_id == "` + greenhouseID + `")`
	}
	if nodeID != "" {
		q += ` |> filter(fn: (r) => r.node_id == "` + nodeID + `")`
	}
	q += ` |> sort(columns: ["_time"], desc: false)`
	q += ` |> group(columns: ["greenhouse_id", "node_id", "_field"])
	  |> keep(columns: ["_time", "greenhouse_id", "node_id", "_field", "_value"])`

	queryAPI := i.client.QueryAPI(i.org)
	result, err := queryAPI.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	type nodeKey struct {
		GreenhouseID, NodeID string
		Time                 time.Time
	}
	// Map: nodeKey -> field -> value
	allMap := make(map[nodeKey]map[string]float64)
	for result.Next() {
		gID := result.Record().ValueByKey("greenhouse_id")
		nID := result.Record().ValueByKey("node_id")
		t := result.Record().Time()
		field := result.Record().Field()
		value, ok := result.Record().Value().(float64)
		if !ok {
			continue
		}
		key := nodeKey{fmt.Sprint(gID), fmt.Sprint(nID), t}
		if _, ok := allMap[key]; !ok {
			allMap[key] = make(map[string]float64)
		}
		allMap[key][field] = value
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var out []models.AverageResult
	for key, fields := range allMap {
		out = append(out, models.AverageResult{
			GreenhouseID: key.GreenhouseID,
			NodeID:       key.NodeID,
			S1Average:    fields["s1_average"],
			S2Average:    fields["s2_average"],
			S3Average:    fields["s3_average"],
			S4Average:    fields["s4_average"],
			S5Average:    fields["s5_average"],
			S6Average:    fields["s6_average"],
			S7Average:    fields["s7_average"],
			S8Average:    fields["s8_average"],
			S9Average:    fields["s9_average"],
			// Timestamp: key.Time,
		})
	}
	return out, nil
}
