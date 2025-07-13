package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"iot-agriculture-backend/internal/services"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Time    string `json:"timestamp"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Time    string      `json:"timestamp"`
}

// sendError sends a standardized error response
func sendError(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    statusCode,
		Message: message,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// sendSuccess sends a standardized success response
func sendSuccess(w http.ResponseWriter, data interface{}, message string) {
	response := SuccessResponse{
		Status:  "success",
		Data:    data,
		Message: message,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next(w, r)
	}
}

// RateLimitMiddleware adds basic rate limiting (placeholder for production)
// This is now replaced by the Redis-based rate limiter in services/rate_limiter.go
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is deprecated - use Redis-based rate limiter instead
		next(w, r)
	}
}

// MonitoringMiddleware adds request monitoring
func MonitoringMiddleware(metricsService *services.MetricsService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next(responseWriter, r)

			// Record metrics
			duration := time.Since(start)
			endpoint := r.URL.Path
			method := r.Method
			status := strconv.Itoa(responseWriter.statusCode)

			if metricsService != nil {
				metricsService.RecordAPIRequest(method, endpoint, status, duration)
			}
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
