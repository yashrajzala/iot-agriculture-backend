package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"iot-agriculture-backend/internal/models"
)

// AveragingService handles sensor data averaging calculations
type AveragingService struct {
	mu      sync.Mutex
	buffers map[string]*models.SensorAverages // key: greenhouse_id|node_id
}

// NewAveragingService creates a new averaging service
func NewAveragingService() *AveragingService {
	return &AveragingService{
		buffers: make(map[string]*models.SensorAverages),
	}
}

// AddSensorData adds sensor data to the averaging system
func (a *AveragingService) AddSensorData(data models.ESP32SensorData) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := data.GreenhouseID + "|" + data.NodeID
	buf, ok := a.buffers[key]
	if !ok {
		buf = &models.SensorAverages{
			GreenhouseID: data.GreenhouseID,
			NodeID:       data.NodeID,
			StartTime:    time.Now(),
		}
		a.buffers[key] = buf
	}
	buf.S1Values = append(buf.S1Values, data.S1)
	buf.S2Values = append(buf.S2Values, data.S2)
	buf.S3Values = append(buf.S3Values, data.S3)
	buf.S4Values = append(buf.S4Values, data.S4)
	buf.S5Values = append(buf.S5Values, data.S5)
	buf.S6Values = append(buf.S6Values, data.S6)
	buf.S7Values = append(buf.S7Values, data.S7)
	buf.S8Values = append(buf.S8Values, data.S8)
	buf.S9Values = append(buf.S9Values, data.S9)
}

// CalculateAndDisplayAverages calculates and displays 60-second averages for all nodes
func (a *AveragingService) CalculateAndDisplayAverages() {
	a.CalculateAndDisplayAveragesWithLogging(nil, nil)
}

// CalculateAndDisplayAveragesWithLogging calculates, displays, and logs 60-second averages for all nodes
func (a *AveragingService) CalculateAndDisplayAveragesWithLogging(influxService *InfluxDBService, metricsService *MetricsService) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.buffers) == 0 {
		fmt.Printf("No sensor data to average in this period.\n")
		return
	}
	for _, buf := range a.buffers {
		result := calculateAveragesForBuffer(buf)
		displayAveragesForResult(result)
		if influxService != nil && influxService.IsConnected() && result.Readings > 0 {
			if err := influxService.LogAverages(result); err != nil {
				fmt.Printf("Warning: Failed to log to InfluxDB: %v\n", err)
				if metricsService != nil {
					metricsService.IncrementInfluxDBWriteErrors()
				}
			} else {
				if metricsService != nil {
					metricsService.IncrementInfluxDBWrites()
				}
			}
		} else if result.Readings == 0 {
			fmt.Printf("Skipping InfluxDB log - no sensor readings for %s/%s in this period\n", buf.GreenhouseID, buf.NodeID)
		}
	}
	// Clear all buffers for next period
	a.buffers = make(map[string]*models.SensorAverages)
}

// GetAverages returns the current averages for all nodes
func (a *AveragingService) GetAverages() []models.AverageResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	results := make([]models.AverageResult, 0, len(a.buffers))
	for _, buf := range a.buffers {
		results = append(results, calculateAveragesForBuffer(buf))
	}
	return results
}

// calculateAveragesForBuffer calculates the averages for a single node buffer
func calculateAveragesForBuffer(buf *models.SensorAverages) models.AverageResult {
	duration := time.Since(buf.StartTime)
	return models.AverageResult{
		GreenhouseID: buf.GreenhouseID,
		NodeID:       buf.NodeID,
		Duration:     duration.Seconds(),
		Readings:     len(buf.S1Values),
		S1Average:    calculateAverage(buf.S1Values),
		S2Average:    calculateAverage(buf.S2Values),
		S3Average:    calculateAverage(buf.S3Values),
		S4Average:    calculateAverage(buf.S4Values),
		S5Average:    calculateAverage(buf.S5Values),
		S6Average:    calculateAverage(buf.S6Values),
		S7Average:    calculateAverage(buf.S7Values),
		S8Average:    calculateAverage(buf.S8Values),
		S9Average:    calculateAverage(buf.S9Values),
	}
}

// displayAveragesForResult displays the calculated averages for a single node
func displayAveragesForResult(result models.AverageResult) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("🕐 60-SECOND SENSOR AVERAGES\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("⏱️  Duration: %.1f seconds\n", result.Duration)
	fmt.Printf("🏠  Greenhouse: %s\n", result.GreenhouseID)
	fmt.Printf("📡  Node: %s\n", result.NodeID)
	fmt.Printf("📊  Total Readings: %d\n", result.Readings)
	fmt.Printf(strings.Repeat("-", 60) + "\n")
	fmt.Printf("🌡️  S1 (Temperature): %.2f\n", result.S1Average)
	fmt.Printf("💧  S2 (Humidity): %.2f\n", result.S2Average)
	fmt.Printf("🌱  S3 (Soil Moisture): %.2f\n", result.S3Average)
	fmt.Printf("💡  S4 (Light): %.2f\n", result.S4Average)
	fmt.Printf("🌿  S5 (CO2): %.2f\n", result.S5Average)
	fmt.Printf("🌪️  S6 (Air Flow): %.2f\n", result.S6Average)
	fmt.Printf("🔋  S7 (Battery): %.2f\n", result.S7Average)
	fmt.Printf("📶  S8 (Signal): %.2f\n", result.S8Average)
	fmt.Printf("⚡  S9 (Power): %.2f\n", result.S9Average)
	fmt.Printf(strings.Repeat("=", 60) + "\n\n")

	// Debug info
	if result.Readings == 0 {
		fmt.Printf("⚠️  WARNING: No sensor readings received in the last 60 seconds for this node!\n")
		fmt.Printf("   Check if ESP32 is sending data to topic for this node\n")
		fmt.Printf("   Check MQTT broker connectivity\n")
		fmt.Printf("   This period will NOT be logged to InfluxDB\n")
	}
}

// (resetAverages is no longer needed; buffers are cleared in CalculateAndDisplayAveragesWithLogging)

// GetReadingCount returns the current number of readings
func (a *AveragingService) GetReadingCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	count := 0
	for _, buf := range a.buffers {
		count += len(buf.S1Values)
	}
	return count
}

// GetDuration returns the current duration since last reset
func (a *AveragingService) GetDuration() time.Duration {
	// Not meaningful for multi-node; return 0
	return 0
}

// calculateAverage calculates the average of a slice of integers
func calculateAverage(values []int) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0
	for _, v := range values {
		sum += v
	}

	result := float64(sum) / float64(len(values))

	return result
}
