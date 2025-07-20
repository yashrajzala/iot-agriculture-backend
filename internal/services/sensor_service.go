package services

import (
	"context"
	"encoding/json"
	"fmt"

	"iot-agriculture-backend/internal/config"
	"iot-agriculture-backend/internal/models"
)

// SensorService handles sensor data processing
type SensorService struct {
	averagingService *AveragingService
	influxService    *InfluxDBService
	metricsService   *MetricsService
	config           *config.Config
}

// NewSensorService creates a new sensor service
func NewSensorService(cfg *config.Config) *SensorService {
	return &SensorService{
		averagingService: NewAveragingService(),
		influxService:    NewInfluxDBService(&cfg.InfluxDB),
		metricsService:   NewMetricsService(),
		config:           cfg,
	}
}

// ProcessSensorData processes incoming sensor data
func (s *SensorService) ProcessSensorData(ctx context.Context, topic string, payload []byte) {
	// Increment MQTT messages metric
	s.metricsService.IncrementMQTTMessages()

	// Remove per-message logging
	// fmt.Printf("MQTT: Received sensor data from %s\n", topic)

	var data models.ESP32SensorData
	if err := json.Unmarshal(payload, &data); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		fmt.Printf("Raw payload: %s\n", string(payload))
		return
	}

	// Add to averaging service
	s.averagingService.AddSensorData(data)

	// Increment sensor readings metric
	s.metricsService.IncrementSensorReadings()
}

// CalculateAndDisplayAverages delegates to the averaging service with InfluxDB logging
func (s *SensorService) CalculateAndDisplayAverages() {
	s.averagingService.CalculateAndDisplayAveragesWithLogging(s.influxService, s.metricsService)
	// Increment sensor averages metric
	s.metricsService.IncrementSensorAverages()
}

// GetInfluxDBService returns the InfluxDB service for external access
func (s *SensorService) GetInfluxDBService() *InfluxDBService {
	return s.influxService
}

// GetAveragingService returns the averaging service for external access
func (s *SensorService) GetAveragingService() *AveragingService {
	return s.averagingService
}

// GetMetricsService returns the metrics service for external access
func (s *SensorService) GetMetricsService() *MetricsService {
	return s.metricsService
}

// Close closes all services
func (s *SensorService) Close() {
	if s.influxService != nil {
		s.influxService.Close()
	}
}
