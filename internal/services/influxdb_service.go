package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"iot-agriculture-backend/internal/models"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// InfluxDBService handles InfluxDB operations
type InfluxDBService struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	org      string
	bucket   string
}

// NewInfluxDBService creates a new InfluxDB service
func NewInfluxDBService() *InfluxDBService {
	// InfluxDB configuration
	url := "http://localhost:8086"
	token := "sR5sjCdApIph5swrk-wKJdJKTyGN20pOhIPrwI3OVUhHtkQD-N8VnPs6hASE7fS2Rajocv17Edh5hOIgT-Lerg=="
	org := "iot-agriculture" // Use your created organization
	bucket := "sensor_data"

	// Create client
	client := influxdb2.NewClient(url, token)
	defer client.Close()

	// Create write API
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// Test connection
	_, err := client.Ping(context.Background())
	if err != nil {
		log.Printf("Warning: Could not connect to InfluxDB: %v", err)
		log.Printf("InfluxDB logging will be disabled")
		return &InfluxDBService{
			client:   nil,
			writeAPI: nil,
			org:      org,
			bucket:   bucket,
		}
	}

	log.Printf("Successfully connected to InfluxDB at %s", url)
	log.Printf("Using organization: %s, bucket: %s", org, bucket)
	return &InfluxDBService{
		client:   client,
		writeAPI: writeAPI,
		org:      org,
		bucket:   bucket,
	}
}

// LogAverages logs sensor averages to InfluxDB
func (i *InfluxDBService) LogAverages(averages models.AverageResult) error {
	if i.client == nil || i.writeAPI == nil {
		return fmt.Errorf("InfluxDB not connected")
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
		return fmt.Errorf("failed to write to InfluxDB: %w", err)
	}

	log.Printf("Logged sensor averages to InfluxDB: %s/%s (%.1fs, %d readings)",
		averages.GreenhouseID, averages.NodeID, averages.Duration, averages.Readings)
	return nil
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
	return i.client != nil && i.writeAPI != nil
}

// GetConnectionInfo returns connection information
func (i *InfluxDBService) GetConnectionInfo() string {
	if i.IsConnected() {
		return fmt.Sprintf("Connected to InfluxDB - Org: %s, Bucket: %s", i.org, i.bucket)
	}
	return "InfluxDB not connected"
}
