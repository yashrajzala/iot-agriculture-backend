package config

import (
	"fmt"
	"log"
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

// InfluxDBConfig holds InfluxDB configuration
type InfluxDBConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

// APIConfig holds API server configuration
type APIConfig struct {
	Port string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// Config holds all application configuration
type Config struct {
	MQTT     MQTTConfig
	Database DatabaseConfig
	InfluxDB InfluxDBConfig
	API      APIConfig
	Redis    RedisConfig
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	config := &Config{
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
		InfluxDB: InfluxDBConfig{
			URL:    getEnv("INFLUXDB_URL", "http://localhost:8086"),
			Token:  getEnv("INFLUXDB_TOKEN", "sR5sjCdApIph5swrk-wKJdJKTyGN20pOhIPrwI3OVUhHtkQD-N8VnPs6hASE7fS2Rajocv17Edh5hOIgT-Lerg=="),
			Org:    getEnv("INFLUXDB_ORG", "iot-agriculture"),
			Bucket: getEnv("INFLUXDB_BUCKET", "sensor_data"),
		},
		API: APIConfig{
			Port: getEnv("API_PORT", "8080"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}

	// Validate critical configuration
	config.validate()
	return config
}

// validate validates critical configuration values
func (c *Config) validate() {
	if c.MQTT.Broker == "" {
		log.Fatal("MQTT_BROKER environment variable is required")
	}
	if c.MQTT.Topic == "" {
		log.Fatal("MQTT_TOPIC environment variable is required")
	}
	// Note: INFLUXDB_TOKEN is optional - service will disable logging if not provided
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
