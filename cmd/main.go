package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iot-agriculture-backend/internal/config"
	"iot-agriculture-backend/internal/mqtt"
	"iot-agriculture-backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting IoT Agriculture Backend with config: %s", cfg.MQTT.String())

	// Create sensor service
	sensorService := services.NewSensorService()
	defer sensorService.Close()

	// Log InfluxDB connection status
	influxService := sensorService.GetInfluxDBService()
	if influxService != nil {
		log.Printf("InfluxDB Status: %s", influxService.GetConnectionInfo())
	}

	// Create MQTT client with message handler
	mqttClient, err := mqtt.NewClient(&cfg.MQTT, sensorService.ProcessSensorData)
	if err != nil {
		log.Fatalf("Failed to create MQTT client: %v", err)
	}
	defer mqttClient.Disconnect()

	// Subscribe to MQTT topic
	if err := mqttClient.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to MQTT topic: %v", err)
	}

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
	log.Println("API server disabled.")

	// Main event loop
	for {
		select {
		case <-ticker.C:
			// Calculate and display 60-second averages
			sensorService.CalculateAndDisplayAverages()

		case <-sigChan:
			// Graceful shutdown
			log.Println("Shutting down gracefully...")
			return

		case <-ctx.Done():
			return
		}
	}
}
