package api

import (
	"fmt"
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
// Supports:
// - Listing all node averages
// - Filtering by greenhouse_id and/or node_id
// - Selecting specific sensors with the 'sensors' query param
func (h *SensorAveragesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Validate query parameters
	if err := h.validateQueryParams(r); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get query parameters
	sensors := r.URL.Query().Get("sensors") // e.g., "S1,S2,S3" or "all"
	greenhouseID := r.URL.Query().Get("greenhouse_id")
	nodeID := r.URL.Query().Get("node_id")

	// Get current averages from the service (now a slice)
	allAverages := h.sensorService.GetAveragingService().GetAverages()

	results := make([]map[string]interface{}, 0)
	for _, averages := range allAverages {
		// Filter by greenhouse_id if specified
		if greenhouseID != "" && averages.GreenhouseID != greenhouseID {
			continue
		}
		// Filter by node_id if specified
		if nodeID != "" && averages.NodeID != nodeID {
			continue
		}

		response := map[string]interface{}{
			"greenhouse_id": averages.GreenhouseID,
			"node_id":       averages.NodeID,
			"duration":      averages.Duration,
			"readings":      averages.Readings,
			"timestamp":     time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			"sensors":       make(map[string]interface{}),
		}

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
			response["sensors"] = sensorMap
		} else {
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

		results = append(results, response)
	}

	if len(results) == 0 {
		sendError(w, http.StatusNotFound, "No sensor averages found for the specified criteria")
		return
	}

	sendSuccess(w, results, "Sensor averages retrieved successfully")
}

// validateQueryParams validates query parameters
func (h *SensorAveragesHandler) validateQueryParams(r *http.Request) error {
	sensors := r.URL.Query().Get("sensors")

	if sensors != "" && sensors != "all" {
		validSensors := []string{"S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8", "S9"}
		requested := strings.Split(sensors, ",")

		for _, s := range requested {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}

			valid := false
			for _, validSensor := range validSensors {
				if s == validSensor {
					valid = true
					break
				}
			}

			if !valid {
				return fmt.Errorf("invalid sensor: %s", s)
			}
		}
	}

	return nil
}

// SensorAveragesLatestHandler handles latest averages from DB
func (h *SensorAveragesHandler) HandleLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if err := h.validateQueryParams(r); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	sensors := r.URL.Query().Get("sensors")
	greenhouseID := r.URL.Query().Get("greenhouse_id")
	nodeID := r.URL.Query().Get("node_id")
	averages, err := h.sensorService.GetInfluxDBService().GetLatestAveragesFromDB(greenhouseID, nodeID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	results := make([]map[string]interface{}, 0)
	for _, avg := range averages {
		response := map[string]interface{}{
			"greenhouse_id": avg.GreenhouseID,
			"node_id":       avg.NodeID,
			"sensors":       make(map[string]interface{}),
		}
		sensorMap := map[string]float64{
			"S1": avg.S1Average,
			"S2": avg.S2Average,
			"S3": avg.S3Average,
			"S4": avg.S4Average,
			"S5": avg.S5Average,
			"S6": avg.S6Average,
			"S7": avg.S7Average,
			"S8": avg.S8Average,
			"S9": avg.S9Average,
		}
		if sensors == "" || sensors == "all" {
			response["sensors"] = sensorMap
		} else {
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
		results = append(results, response)
	}
	if len(results) == 0 {
		sendError(w, http.StatusNotFound, "No sensor averages found for the specified criteria")
		return
	}
	sendSuccess(w, results, "Latest sensor averages retrieved from database")
}

// SensorAveragesAllHandler handles fetching all average data from DB
func (h *SensorAveragesHandler) HandleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if err := h.validateQueryParams(r); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	sensors := r.URL.Query().Get("sensors")
	greenhouseID := r.URL.Query().Get("greenhouse_id")
	nodeID := r.URL.Query().Get("node_id")
	averages, err := h.sensorService.GetInfluxDBService().GetAllAveragesFromDB(greenhouseID, nodeID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	results := make([]map[string]interface{}, 0)
	for _, avg := range averages {
		response := map[string]interface{}{
			"greenhouse_id": avg.GreenhouseID,
			"node_id":       avg.NodeID,
			"sensors":       make(map[string]interface{}),
		}
		sensorMap := map[string]float64{
			"S1": avg.S1Average,
			"S2": avg.S2Average,
			"S3": avg.S3Average,
			"S4": avg.S4Average,
			"S5": avg.S5Average,
			"S6": avg.S6Average,
			"S7": avg.S7Average,
			"S8": avg.S8Average,
			"S9": avg.S9Average,
		}
		if sensors == "" || sensors == "all" {
			response["sensors"] = sensorMap
		} else {
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
		results = append(results, response)
	}
	if len(results) == 0 {
		sendError(w, http.StatusNotFound, "No sensor averages found for the specified criteria")
		return
	}
	sendSuccess(w, results, "All sensor averages retrieved from database")
}
