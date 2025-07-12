package api

import (
	"encoding/json"
	"net/http"
	"time"

	"iot-agriculture-backend/internal/services"
)

// DatabaseHealthHandler handles database health check requests
type DatabaseHealthHandler struct {
	sensorService *services.SensorService
}

// NewDatabaseHealthHandler creates a new database health handler
func NewDatabaseHealthHandler(sensorService *services.SensorService) *DatabaseHealthHandler {
	return &DatabaseHealthHandler{
		sensorService: sensorService,
	}
}

// Handle handles database health check requests
func (h *DatabaseHealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get InfluxDB service
	influxService := h.sensorService.GetInfluxDBService()

	// Check database health
	health := map[string]interface{}{
		"status":    "unknown",
		"connected": false,
		"message":   "Database service not available",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if influxService != nil {
		isConnected := influxService.IsConnected()
		connectionInfo := influxService.GetConnectionInfo()

		health["status"] = "ok"
		if isConnected {
			health["status"] = "connected"
		} else {
			health["status"] = "disconnected"
		}
		health["connected"] = isConnected
		health["message"] = connectionInfo
	}

	// Return JSON response
	json.NewEncoder(w).Encode(health)
}
