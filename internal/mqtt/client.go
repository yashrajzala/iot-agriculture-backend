package mqtt

import (
	"fmt"
	"log"
	"time"

	"iot-agriculture-backend/internal/config"
	"iot-agriculture-backend/internal/services"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler is a function type for handling MQTT messages
type MessageHandler func(topic string, payload []byte)

// Client wraps the MQTT client with additional functionality
type Client struct {
	client         MQTT.Client
	config         *config.MQTTConfig
	handler        MessageHandler
	metricsService *services.MetricsService
}

// NewClient creates a new MQTT client
func NewClient(cfg *config.MQTTConfig, handler MessageHandler, metricsService *services.MetricsService) (*Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Broker, cfg.Port))
	opts.SetClientID(cfg.ClientID)
	opts.SetConnectTimeout(30 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(1 * time.Minute)
	opts.SetKeepAlive(60 * time.Second)   // Send keep-alive every 60 seconds
	opts.SetPingTimeout(10 * time.Second) // Wait 10 seconds for ping response
	opts.SetCleanSession(true)            // Start with clean session
	opts.SetResumeSubs(true)              // Resume subscriptions after reconnect

	// Note: We use topic-specific handlers instead of default handler

	// Set connection lost handler
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
		log.Printf("Attempting to reconnect...")
		if metricsService != nil {
			metricsService.SetMQTTConnectionStatus(false)
			metricsService.IncrementMQTTReconnections()
		}
	})

	// Set on connect handler
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		log.Printf("Connected to MQTT broker: %s", cfg.String())
		log.Printf("Client ID: %s, Keep-alive: 60s", cfg.ClientID)
		if metricsService != nil {
			metricsService.SetMQTTConnectionStatus(true)
		}
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return &Client{
		client:         client,
		config:         cfg,
		handler:        handler,
		metricsService: metricsService,
	}, nil
}

// Subscribe subscribes to the configured topic
func (c *Client) Subscribe() error {
	if token := c.client.Subscribe(c.config.Topic, 1, func(client MQTT.Client, msg MQTT.Message) {
		// Check for empty or null payloads
		if len(msg.Payload()) == 0 {
			log.Printf("WARNING: Empty MQTT payload received!")
			return // Don't process empty messages
		}

		if c.handler != nil {
			c.handler(msg.Topic(), msg.Payload())
		}
	}); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", c.config.Topic, token.Error())
	}
	log.Printf("Subscribed to topic: %s", c.config.Topic)
	return nil
}

// Disconnect disconnects from the MQTT broker
func (c *Client) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
		log.Println("Disconnected from MQTT broker")
	}
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.client != nil && c.client.IsConnected()
}

// GetConnectionInfo returns connection information
func (c *Client) GetConnectionInfo() string {
	if c.client == nil {
		return "MQTT client not initialized"
	}

	if c.client.IsConnected() {
		return fmt.Sprintf("Connected to MQTT broker: %s", c.config.String())
	}
	return fmt.Sprintf("Disconnected from MQTT broker: %s", c.config.String())
}
