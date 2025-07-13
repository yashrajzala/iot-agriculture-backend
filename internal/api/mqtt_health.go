package api

import (
	"net/http"
	"time"

	"iot-agriculture-backend/internal/mqtt"
	"iot-agriculture-backend/internal/services"
)

// MQTTHealthHandler handles MQTT connection health check requests
type MQTTHealthHandler struct {
	sensorService *services.SensorService
	mqttClient    *mqtt.Client
}

// NewMQTTHealthHandler creates a new MQTT health handler
func NewMQTTHealthHandler(sensorService *services.SensorService, mqttClient *mqtt.Client) *MQTTHealthHandler {
	return &MQTTHealthHandler{
		sensorService: sensorService,
		mqttClient:    mqttClient,
	}
}

// Handle handles MQTT connection health check requests
func (h *MQTTHealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check MQTT connection health
	health := map[string]interface{}{
		"status":    "unknown",
		"connected": false,
		"message":   "MQTT client not available",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if h.mqttClient != nil {
		isConnected := h.mqttClient.IsConnected()
		connectionInfo := h.mqttClient.GetConnectionInfo()

		health["status"] = "ok"
		if isConnected {
			health["status"] = "connected"
		} else {
			health["status"] = "disconnected"
		}
		health["connected"] = isConnected
		health["message"] = connectionInfo
	}

	// Return success response
	sendSuccess(w, health, "MQTT health check completed")
}
