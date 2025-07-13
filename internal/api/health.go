package api

import (
	"encoding/json"
	"net/http"
	"time"

	"iot-agriculture-backend/internal/mqtt"
	"iot-agriculture-backend/internal/services"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	sensorService *services.SensorService
	mqttClient    *mqtt.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(sensorService *services.SensorService, mqttClient *mqtt.Client) *HealthHandler {
	return &HealthHandler{
		sensorService: sensorService,
		mqttClient:    mqttClient,
	}
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceInfo `json:"services"`
}

// ServiceInfo represents individual service health
type ServiceInfo struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// Handle handles health check requests
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check all services
	services := make(map[string]ServiceInfo)

	// Check MQTT connection
	mqttStatus := "healthy"
	mqttMessage := "Connected to MQTT broker"
	if h.mqttClient == nil || !h.mqttClient.IsConnected() {
		mqttStatus = "unhealthy"
		mqttMessage = "MQTT client not connected"
	}
	services["mqtt"] = ServiceInfo{
		Status:    mqttStatus,
		Message:   mqttMessage,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check InfluxDB connection
	influxService := h.sensorService.GetInfluxDBService()
	influxStatus := "healthy"
	influxMessage := "Connected to InfluxDB"
	if influxService == nil || !influxService.IsConnected() {
		influxStatus = "unhealthy"
		influxMessage = "InfluxDB not connected"
	}
	services["influxdb"] = ServiceInfo{
		Status:    influxStatus,
		Message:   influxMessage,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check averaging service
	avgService := h.sensorService.GetAveragingService()
	avgStatus := "healthy"
	avgMessage := "Averaging service operational"
	if avgService == nil {
		avgStatus = "unhealthy"
		avgMessage = "Averaging service not initialized"
	}
	services["averaging"] = ServiceInfo{
		Status:    avgStatus,
		Message:   avgMessage,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check metrics service
	metricsService := h.sensorService.GetMetricsService()
	metricsStatus := "healthy"
	metricsMessage := "Metrics service operational"
	if metricsService == nil {
		metricsStatus = "unhealthy"
		metricsMessage = "Metrics service not initialized"
	}
	services["metrics"] = ServiceInfo{
		Status:    metricsStatus,
		Message:   metricsMessage,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Determine overall status
	overallStatus := "healthy"
	httpStatus := http.StatusOK

	for _, service := range services {
		if service.Status == "unhealthy" {
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			break
		}
	}

	// Create health response
	health := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(metricsService.GetStartTime()).String(),
		Version:   "1.0.0",
		Services:  services,
	}

	// Set appropriate HTTP status code
	w.WriteHeader(httpStatus)

	// Send response
	response := SuccessResponse{
		Status:  "success",
		Data:    health,
		Message: "Health check completed",
		Time:    time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
