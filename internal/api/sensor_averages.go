package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"iot-agriculture-backend/internal/services"
)

// SensorAveragesHandler handles sensor averages requests
type SensorAveragesHandler struct {
	sensorService *services.SensorService
}

// NewSensorAveragesHandler creates a new sensor averages handler
func NewSensorAveragesHandler(sensorService *services.SensorService) *SensorAveragesHandler {
	return &SensorAveragesHandler{
		sensorService: sensorService,
	}
}

// Handle handles sensor averages requests
func (h *SensorAveragesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	sensors := r.URL.Query().Get("sensors") // e.g., "S1,S2,S3" or "all"
	greenhouseID := r.URL.Query().Get("greenhouse_id")
	nodeID := r.URL.Query().Get("node_id")

	// Get current averages from the service
	averages := h.sensorService.GetAveragingService().GetAverages()

	// Prepare response
	response := map[string]interface{}{
		"greenhouse_id": averages.GreenhouseID,
		"node_id":       averages.NodeID,
		"duration":      averages.Duration,
		"readings":      averages.Readings,
		"timestamp":     time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"sensors":       make(map[string]interface{}),
	}

	// Determine which sensors to include
	sensorMap := map[string]float64{
		"S1": averages.S1Average,
		"S2": averages.S2Average,
		"S3": averages.S3Average,
		"S4": averages.S4Average,
		"S5": averages.S5Average,
		"S6": averages.S6Average,
		"S7": averages.S7Average,
		"S8": averages.S8Average,
		"S9": averages.S9Average,
	}

	// Filter sensors based on request
	if sensors == "" || sensors == "all" {
		// Return all sensors
		response["sensors"] = sensorMap
	} else {
		// Return only requested sensors
		requestedSensors := strings.Split(sensors, ",")
		filteredSensors := make(map[string]interface{})

		for _, sensor := range requestedSensors {
			sensor = strings.TrimSpace(sensor)
			if value, exists := sensorMap[sensor]; exists {
				filteredSensors[sensor] = value
			}
		}

		response["sensors"] = filteredSensors
	}

	// Filter by greenhouse_id if specified
	if greenhouseID != "" && averages.GreenhouseID != greenhouseID {
		response["error"] = "Greenhouse ID not found"
		response["sensors"] = make(map[string]interface{})
	}

	// Filter by node_id if specified
	if nodeID != "" && averages.NodeID != nodeID {
		response["error"] = "Node ID not found"
		response["sensors"] = make(map[string]interface{})
	}

	// Return JSON response
	json.NewEncoder(w).Encode(response)
}
