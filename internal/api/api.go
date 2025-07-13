package api

import (
	"net/http"
	"time"

	"iot-agriculture-backend/internal/mqtt"
	"iot-agriculture-backend/internal/services"
)

// Server represents the API server
type Server struct {
	sensorService *services.SensorService
	mqttClient    *mqtt.Client
	server        *http.Server
}

// NewServer creates a new API server
func NewServer(sensorService *services.SensorService, mqttClient *mqtt.Client, port string) *Server {
	mux := http.NewServeMux()

	server := &Server{
		sensorService: sensorService,
		mqttClient:    mqttClient,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	// Create handlers
	dbHealthHandler := NewDatabaseHealthHandler(sensorService)
	mqttHealthHandler := NewMQTTHealthHandler(sensorService, mqttClient)
	sensorAveragesHandler := NewSensorAveragesHandler(sensorService)

	// Create monitoring middleware
	monitoringMiddleware := MonitoringMiddleware(sensorService.GetMetricsService())

	// Register routes with enhanced middleware
	mux.HandleFunc("/health/database", SecurityMiddleware(RateLimitMiddleware(monitoringMiddleware(CORSMiddleware(dbHealthHandler.Handle)))))
	mux.HandleFunc("/health/mqtt", SecurityMiddleware(RateLimitMiddleware(monitoringMiddleware(CORSMiddleware(mqttHealthHandler.Handle)))))
	mux.HandleFunc("/sensors/averages", SecurityMiddleware(RateLimitMiddleware(monitoringMiddleware(CORSMiddleware(sensorAveragesHandler.Handle)))))

	// Metrics endpoint (no rate limiting for Prometheus scraping)
	mux.HandleFunc("/metrics", SecurityMiddleware(CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		sensorService.GetMetricsService().GetMetricsHandler().ServeHTTP(w, r)
	})))

	return server
}

// Start starts the API server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop() error {
	return s.server.Close()
}
