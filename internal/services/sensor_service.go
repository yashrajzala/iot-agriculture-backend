package services

import (
	"encoding/json"
	"fmt"

	"iot-agriculture-backend/internal/models"
)

// SensorService handles sensor data processing
type SensorService struct {
	averagingService *AveragingService
	influxService    *InfluxDBService
}

// NewSensorService creates a new sensor service
func NewSensorService() *SensorService {
	return &SensorService{
		averagingService: NewAveragingService(),
		influxService:    NewInfluxDBService(),
	}
}

// ProcessSensorData processes incoming sensor data
func (s *SensorService) ProcessSensorData(topic string, payload []byte) {
	fmt.Printf("\n=== New ESP32 Sensor Data ===\n")
	fmt.Printf("Topic: %s\n", topic)
	fmt.Printf("Payload: %s\n", string(payload))

	var data models.ESP32SensorData
	if err := json.Unmarshal(payload, &data); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		fmt.Printf("Raw payload: %s\n", string(payload))
		return
	}

	// Check for 0 values in sensor data
	zeroCount := 0
	if data.S1 == 0 {
		zeroCount++
	}
	if data.S2 == 0 {
		zeroCount++
	}
	if data.S3 == 0 {
		zeroCount++
	}
	if data.S4 == 0 {
		zeroCount++
	}
	if data.S5 == 0 {
		zeroCount++
	}
	if data.S6 == 0 {
		zeroCount++
	}
	if data.S7 == 0 {
		zeroCount++
	}
	if data.S8 == 0 {
		zeroCount++
	}
	if data.S9 == 0 {
		zeroCount++
	}

	if zeroCount > 0 {
		fmt.Printf("WARNING: ESP32 sent %d zero values! (S1:%d, S2:%d, S3:%d, S4:%d, S5:%d, S6:%d, S7:%d, S8:%d, S9:%d)\n",
			zeroCount, data.S1, data.S2, data.S3, data.S4, data.S5, data.S6, data.S7, data.S8, data.S9)
	}

	// Display individual sensor values
	s.displaySensorData(data)

	// Add to averaging service
	s.averagingService.AddSensorData(data)
}

// displaySensorData displays individual sensor values
func (s *SensorService) displaySensorData(data models.ESP32SensorData) {
	fmt.Printf("Greenhouse: %s\n", data.GreenhouseID)
	fmt.Printf("Node: %s\n", data.NodeID)
	fmt.Printf("S1: %d\n", data.S1)
	fmt.Printf("S2: %d\n", data.S2)
	fmt.Printf("S3: %d\n", data.S3)
	fmt.Printf("S4: %d\n", data.S4)
	fmt.Printf("S5: %d\n", data.S5)
	fmt.Printf("S6: %d\n", data.S6)
	fmt.Printf("S7: %d\n", data.S7)
	fmt.Printf("S8: %d\n", data.S8)
	fmt.Printf("S9: %d\n", data.S9)
	fmt.Printf("============================\n\n")
}

// CalculateAndDisplayAverages delegates to the averaging service with InfluxDB logging
func (s *SensorService) CalculateAndDisplayAverages() {
	s.averagingService.CalculateAndDisplayAveragesWithLogging(s.influxService)
}

// GetInfluxDBService returns the InfluxDB service for external access
func (s *SensorService) GetInfluxDBService() *InfluxDBService {
	return s.influxService
}

// GetAveragingService returns the averaging service for external access
func (s *SensorService) GetAveragingService() *AveragingService {
	return s.averagingService
}

// Close closes all services
func (s *SensorService) Close() {
	if s.influxService != nil {
		s.influxService.Close()
	}
}
