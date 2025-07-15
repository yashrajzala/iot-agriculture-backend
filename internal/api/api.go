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
	rateLimiter   *services.RateLimiter
	server        *http.Server
}

// NewServer creates a new API server
func NewServer(sensorService *services.SensorService, mqttClient *mqtt.Client, rateLimiter *services.RateLimiter, port string) *Server {
	mux := http.NewServeMux()

	server := &Server{
		sensorService: sensorService,
		mqttClient:    mqttClient,
		rateLimiter:   rateLimiter,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	// Create handlers
	healthHandler := NewHealthHandler(sensorService, mqttClient)
	dbHealthHandler := NewDatabaseHealthHandler(sensorService)
	mqttHealthHandler := NewMQTTHealthHandler(sensorService, mqttClient)
	sensorAveragesHandler := NewSensorAveragesHandler(sensorService)

	// Create monitoring middleware
	monitoringMiddleware := MonitoringMiddleware(sensorService.GetMetricsService())

	// Create rate limiting configuration
	rateLimitConfig := services.RateLimitConfig{
		RequestsPerMinute: 60,   // 60 requests per minute
		RequestsPerHour:   1000, // 1000 requests per hour
		BurstSize:         10,   // Allow burst of 10 requests
	}

	// Register routes with enhanced middleware
	mux.HandleFunc("/health", SecurityMiddleware(rateLimiter.RateLimitMiddleware(rateLimitConfig)(monitoringMiddleware(CORSMiddleware(healthHandler.Handle)))))
	mux.HandleFunc("/health/database", SecurityMiddleware(rateLimiter.RateLimitMiddleware(rateLimitConfig)(monitoringMiddleware(CORSMiddleware(dbHealthHandler.Handle)))))
	mux.HandleFunc("/health/mqtt", SecurityMiddleware(rateLimiter.RateLimitMiddleware(rateLimitConfig)(monitoringMiddleware(CORSMiddleware(mqttHealthHandler.Handle)))))
	mux.HandleFunc("/sensors/averages", SecurityMiddleware(rateLimiter.RateLimitMiddleware(rateLimitConfig)(monitoringMiddleware(CORSMiddleware(sensorAveragesHandler.Handle)))))
	mux.HandleFunc("/sensors/averages/latest", SecurityMiddleware(rateLimiter.RateLimitMiddleware(rateLimitConfig)(monitoringMiddleware(CORSMiddleware(sensorAveragesHandler.HandleLatest)))))
	mux.HandleFunc("/sensors/averages/all", SecurityMiddleware(rateLimiter.RateLimitMiddleware(rateLimitConfig)(monitoringMiddleware(CORSMiddleware(sensorAveragesHandler.HandleAll)))))

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
