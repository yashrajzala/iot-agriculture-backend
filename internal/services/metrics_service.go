package services

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsService handles Prometheus metrics
type MetricsService struct {
	// MQTT metrics
	mqttMessagesReceived  prometheus.Counter
	mqttConnectionStatus  prometheus.Gauge
	mqttReconnectionCount prometheus.Counter

	// Sensor metrics
	sensorReadingsProcessed  prometheus.Counter
	sensorAveragesCalculated prometheus.Counter
	sensorZeroValueCount     prometheus.Counter

	// InfluxDB metrics
	influxDBWritesTotal      prometheus.Counter
	influxDBWriteErrors      prometheus.Counter
	influxDBConnectionStatus prometheus.Gauge

	// API metrics
	apiRequestsTotal   *prometheus.CounterVec
	apiRequestDuration *prometheus.HistogramVec

	// System metrics
	uptime    prometheus.Gauge
	startTime time.Time

	mu sync.RWMutex
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	ms := &MetricsService{
		startTime: time.Now(),
	}

	// Initialize MQTT metrics
	ms.mqttMessagesReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mqtt_messages_received_total",
		Help: "Total number of MQTT messages received",
	})

	ms.mqttConnectionStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mqtt_connection_status",
		Help: "MQTT connection status (1 = connected, 0 = disconnected)",
	})

	ms.mqttReconnectionCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mqtt_reconnection_count_total",
		Help: "Total number of MQTT reconnections",
	})

	// Initialize sensor metrics
	ms.sensorReadingsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sensor_readings_processed_total",
		Help: "Total number of sensor readings processed",
	})

	ms.sensorAveragesCalculated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sensor_averages_calculated_total",
		Help: "Total number of sensor averages calculated",
	})

	ms.sensorZeroValueCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sensor_zero_values_total",
		Help: "Total number of zero values received from sensors",
	})

	// Initialize InfluxDB metrics
	ms.influxDBWritesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "influxdb_writes_total",
		Help: "Total number of InfluxDB writes",
	})

	ms.influxDBWriteErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "influxdb_write_errors_total",
		Help: "Total number of InfluxDB write errors",
	})

	ms.influxDBConnectionStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "influxdb_connection_status",
		Help: "InfluxDB connection status (1 = connected, 0 = disconnected)",
	})

	// Initialize API metrics
	ms.apiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	ms.apiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Initialize system metrics
	ms.uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "application_uptime_seconds",
		Help: "Application uptime in seconds",
	})

	// Register all metrics
	prometheus.MustRegister(
		ms.mqttMessagesReceived,
		ms.mqttConnectionStatus,
		ms.mqttReconnectionCount,
		ms.sensorReadingsProcessed,
		ms.sensorAveragesCalculated,
		ms.sensorZeroValueCount,
		ms.influxDBWritesTotal,
		ms.influxDBWriteErrors,
		ms.influxDBConnectionStatus,
		ms.apiRequestsTotal,
		ms.apiRequestDuration,
		ms.uptime,
	)

	// Start uptime updater
	go ms.updateUptime()

	return ms
}

// updateUptime updates the uptime metric
func (ms *MetricsService) updateUptime() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ms.uptime.Set(time.Since(ms.startTime).Seconds())
	}
}

// MQTT Metrics
func (ms *MetricsService) IncrementMQTTMessages() {
	ms.mqttMessagesReceived.Inc()
}

func (ms *MetricsService) SetMQTTConnectionStatus(connected bool) {
	if connected {
		ms.mqttConnectionStatus.Set(1)
	} else {
		ms.mqttConnectionStatus.Set(0)
	}
}

func (ms *MetricsService) IncrementMQTTReconnections() {
	ms.mqttReconnectionCount.Inc()
}

// Sensor Metrics
func (ms *MetricsService) IncrementSensorReadings() {
	ms.sensorReadingsProcessed.Inc()
}

func (ms *MetricsService) IncrementSensorAverages() {
	ms.sensorAveragesCalculated.Inc()
}

func (ms *MetricsService) IncrementSensorZeroValues(count int) {
	ms.sensorZeroValueCount.Add(float64(count))
}

// InfluxDB Metrics
func (ms *MetricsService) IncrementInfluxDBWrites() {
	ms.influxDBWritesTotal.Inc()
}

func (ms *MetricsService) IncrementInfluxDBWriteErrors() {
	ms.influxDBWriteErrors.Inc()
}

func (ms *MetricsService) SetInfluxDBConnectionStatus(connected bool) {
	if connected {
		ms.influxDBConnectionStatus.Set(1)
	} else {
		ms.influxDBConnectionStatus.Set(0)
	}
}

// API Metrics
func (ms *MetricsService) RecordAPIRequest(method, endpoint, status string, duration time.Duration) {
	ms.apiRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	ms.apiRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// GetMetricsHandler returns the Prometheus metrics handler
func (ms *MetricsService) GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}
