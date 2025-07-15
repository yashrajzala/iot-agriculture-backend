package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iot-agriculture-backend/internal/api"
	"iot-agriculture-backend/internal/config"
	"iot-agriculture-backend/internal/mqtt"
	"iot-agriculture-backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting IoT Agriculture Backend with config: %s", cfg.MQTT.String())

	// Create sensor service
	sensorService := services.NewSensorService(cfg)
	defer sensorService.Close()

	// Log InfluxDB connection status
	influxService := sensorService.GetInfluxDBService()
	if influxService != nil {
		log.Printf("InfluxDB Status: %s", influxService.GetConnectionInfo())
	}

	// Create MQTT client with message handler and metrics
	mqttClient, err := mqtt.NewClient(&cfg.MQTT, sensorService.ProcessSensorData, sensorService.GetMetricsService())
	if err != nil {
		log.Fatalf("Failed to create MQTT client: %v", err)
	}
	defer mqttClient.Disconnect()

	// Subscribe to MQTT topic
	if err := mqttClient.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to MQTT topic: %v", err)
	}

	// Create rate limiter
	rateLimiter := services.NewRateLimiter(cfg.Redis.URL)
	defer rateLimiter.Close()

	// Create API server
	apiServer := api.NewServer(sensorService, mqttClient, rateLimiter, "8080")

	// Start API server in a goroutine
	go func() {
		log.Printf("Starting API server on port 8080")
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	// Start averaging timer
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("IoT Agriculture Backend started. Press Ctrl+C to stop.")
	log.Println("MQTT data processing and 60-second averaging enabled.")
	log.Println("API server enabled on port 8080.")

	// Main event loop
	for {
		select {
		case <-ticker.C:
			// Calculate and display 60-second averages
			sensorService.CalculateAndDisplayAverages()

		case <-sigChan:
			// Graceful shutdown
			log.Println("Shutting down gracefully...")

			// Stop the ticker first to prevent new averaging calculations
			ticker.Stop()

			// Stop API server
			apiServer.Stop()

			// Close services (this will cancel the InfluxDB context)
			sensorService.Close()

			// Small delay to ensure cleanup completes
			time.Sleep(200 * time.Millisecond)

			log.Println("Shutdown completed")
			return

		case <-ctx.Done():
			return
		}
	}
}
