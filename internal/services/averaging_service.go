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
	averages *models.SensorAverages
	pool     *sync.Pool // Memory pool for sensor data slices
}

// NewAveragingService creates a new averaging service
func NewAveragingService() *AveragingService {
	return &AveragingService{
		averages: &models.SensorAverages{
			StartTime: time.Now(),
		},
		pool: nil, // Not using pool for now
	}
}

// AddSensorData adds sensor data to the averaging system
func (a *AveragingService) AddSensorData(data models.ESP32SensorData) {
	a.averages.GreenhouseID = data.GreenhouseID
	a.averages.NodeID = data.NodeID
	a.averages.S1Values = append(a.averages.S1Values, data.S1)
	a.averages.S2Values = append(a.averages.S2Values, data.S2)
	a.averages.S3Values = append(a.averages.S3Values, data.S3)
	a.averages.S4Values = append(a.averages.S4Values, data.S4)
	a.averages.S5Values = append(a.averages.S5Values, data.S5)
	a.averages.S6Values = append(a.averages.S6Values, data.S6)
	a.averages.S7Values = append(a.averages.S7Values, data.S7)
	a.averages.S8Values = append(a.averages.S8Values, data.S8)
	a.averages.S9Values = append(a.averages.S9Values, data.S9)
}

// CalculateAndDisplayAverages calculates and displays 60-second averages
func (a *AveragingService) CalculateAndDisplayAverages() {
	result := a.calculateAverages()
	a.displayAverages(result)
	a.resetAverages()
}

// CalculateAndDisplayAveragesWithLogging calculates, displays, and logs 60-second averages
func (a *AveragingService) CalculateAndDisplayAveragesWithLogging(influxService *InfluxDBService, metricsService *MetricsService) {
	result := a.calculateAverages()
	a.displayAverages(result)

	// Only log to InfluxDB if we have actual readings (not 0)
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
		fmt.Printf("Skipping InfluxDB log - no sensor readings in this period\n")
	}

	a.resetAverages()
}

// GetAverages returns the current averages without displaying them
func (a *AveragingService) GetAverages() models.AverageResult {
	return a.calculateAverages()
}

// calculateAverages calculates the averages for all sensors
func (a *AveragingService) calculateAverages() models.AverageResult {
	duration := time.Since(a.averages.StartTime)

	return models.AverageResult{
		GreenhouseID: a.averages.GreenhouseID,
		NodeID:       a.averages.NodeID,
		Duration:     duration.Seconds(),
		Readings:     len(a.averages.S1Values),
		S1Average:    calculateAverage(a.averages.S1Values),
		S2Average:    calculateAverage(a.averages.S2Values),
		S3Average:    calculateAverage(a.averages.S3Values),
		S4Average:    calculateAverage(a.averages.S4Values),
		S5Average:    calculateAverage(a.averages.S5Values),
		S6Average:    calculateAverage(a.averages.S6Values),
		S7Average:    calculateAverage(a.averages.S7Values),
		S8Average:    calculateAverage(a.averages.S8Values),
		S9Average:    calculateAverage(a.averages.S9Values),
	}
}

// displayAverages displays the calculated averages
func (a *AveragingService) displayAverages(result models.AverageResult) {
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
		fmt.Printf("⚠️  WARNING: No sensor readings received in the last 60 seconds!\n")
		fmt.Printf("   Check if ESP32 is sending data to topic: esp32/data\n")
		fmt.Printf("   Check MQTT broker connectivity\n")
		fmt.Printf("   This period will NOT be logged to InfluxDB\n")
	}
}

// resetAverages resets the averaging system for the next period
func (a *AveragingService) resetAverages() {
	// Clear all slices (don't return to pool, just clear them)
	a.averages.S1Values = a.averages.S1Values[:0]
	a.averages.S2Values = a.averages.S2Values[:0]
	a.averages.S3Values = a.averages.S3Values[:0]
	a.averages.S4Values = a.averages.S4Values[:0]
	a.averages.S5Values = a.averages.S5Values[:0]
	a.averages.S6Values = a.averages.S6Values[:0]
	a.averages.S7Values = a.averages.S7Values[:0]
	a.averages.S8Values = a.averages.S8Values[:0]
	a.averages.S9Values = a.averages.S9Values[:0]

	a.averages.StartTime = time.Now()
}

// GetReadingCount returns the current number of readings
func (a *AveragingService) GetReadingCount() int {
	return len(a.averages.S1Values)
}

// GetDuration returns the current duration since last reset
func (a *AveragingService) GetDuration() time.Duration {
	return time.Since(a.averages.StartTime)
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
