package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// MQTTConfig holds MQTT broker configuration
type MQTTConfig struct {
	Broker   string
	Port     int
	ClientID string
	Topic    string
	Username string
	Password string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// APIConfig holds API server configuration
type APIConfig struct {
	Port string
}

// Config holds all application configuration
type Config struct {
	MQTT     MQTTConfig
	Database DatabaseConfig
	API      APIConfig
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "192.168.20.1"),
			Port:     getEnvAsInt("MQTT_PORT", 1883),
			ClientID: getEnv("MQTT_CLIENT_ID", "go-mqtt-subscriber-"+fmt.Sprintf("%d", time.Now().Unix())),
			Topic:    getEnv("MQTT_TOPIC", "esp32/data"),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "iot_agriculture"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		API: APIConfig{
			Port: getEnv("API_PORT", "8080"),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// String returns a string representation of the MQTT configuration
func (c *MQTTConfig) String() string {
	return fmt.Sprintf("MQTT Broker: %s:%d, Topic: %s, ClientID: %s",
		c.Broker, c.Port, c.Topic, c.ClientID)
}
