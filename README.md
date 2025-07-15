# 🌱 IoT Agriculture Backend

A production-ready Go backend for IoT agriculture systems that processes real-time sensor data from ESP32 devices via MQTT, calculates 60-second averages, stores data in InfluxDB, and provides secure REST APIs for frontend integration. Now supports multi-node, multi-sensor, and database-backed queries.

## 🎯 Features

### 🔄 **Real-time Data Processing**
- **MQTT Sensor Data Processing**: Real-time processing of ESP32 sensor data from 5 nodes
- **60-Second Averaging**: Automatic calculation and logging of sensor averages
- **Clean Logging**: Minimal console output with prominent 60-second averages display

### 🗄️ **Data Storage & Management**
- **InfluxDB Integration**: Efficient time-series data storage with circuit breaker protection
- **Automatic Data Logging**: Every 60 seconds, averages are stored to InfluxDB
- **Connection Resilience**: Auto-reconnection and failure recovery
- **Multi-node Support**: Each node's data is averaged and stored independently

### 🔒 **Security & API**
- **REST API Endpoints**: Health checks and sensor data retrieval with security headers
- **Input Validation**: Query parameter sanitization and validation
- **Security Headers**: XSS protection, content type options, frame options
- **Environment-based Configuration**: No hardcoded secrets
- **Rate Limiting**: Redis-based rate limiting with sliding window algorithm
- **Load Balancer Health Checks**: Comprehensive `/health` endpoint for monitoring

### 📊 **Monitoring & Observability**
- **Prometheus Metrics**: Comprehensive monitoring with `/metrics` endpoint
- **Health Checks**: Database and MQTT connectivity monitoring
- **Real-time Metrics**: MQTT messages, sensor readings, API requests, system uptime
- **Structured Logging**: Clean, informative console output

### ⚡ **Performance & Reliability**
- **Circuit Breaker Pattern**: InfluxDB write protection with automatic recovery
- **Memory Optimization**: Efficient data structures for high-throughput processing
- **Connection Pooling**: Optimized database connections
- **Graceful Shutdown**: Proper cleanup and resource management

## 🏗️ Architecture

```
iot-agriculture-backend/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── api/                       # REST API endpoints
│   │   ├── api.go                 # Main API server setup with security middleware
│   │   ├── middleware.go          # CORS, security, and monitoring middleware
│   │   ├── database_health.go     # Database health check API
│   │   ├── mqtt_health.go         # MQTT connection health API
│   │   ├── sensor_averages.go     # Sensor averages data API with validation
│   │   └── README.md              # API documentation
│   ├── config/                    # Configuration management
│   │   └── config.go              # Environment-based configuration with validation
│   ├── models/                    # Data models
│   │   └── sensor.go              # ESP32 sensor data structures
│   ├── mqtt/                      # MQTT client abstraction
│   │   └── client.go              # MQTT client with auto-reconnection
│   └── services/                  # Business logic services
│       ├── sensor_service.go      # Sensor data processing with clean logging
│       ├── averaging_service.go   # 60-second averaging logic
│       ├── influxdb_service.go    # InfluxDB integration with circuit breaker
│       └── metrics_service.go     # Prometheus metrics collection
├── go.mod                         # Go module dependencies
└── README.md                      # This file
```

## 📋 Prerequisites

- **Go 1.24.5** or higher
- **MQTT Broker** (e.g., Mosquitto, HiveMQ, AWS IoT)
- **InfluxDB 2.x** running on localhost:8086
- **Redis** running on localhost:6379 (for rate limiting)
- **ESP32 Device** publishing sensor data

## 🚀 Quick Start

### 1. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd iot-agriculture-backend
   go mod tidy
   ```

### 2. **Configure Environment Variables**
   ```bash
# Required for InfluxDB logging
export INFLUXDB_TOKEN="your-influxdb-token-here"

# Optional: Customize MQTT settings
export MQTT_BROKER="192.168.20.1"
export MQTT_PORT="1883"
export MQTT_TOPIC="greenhouse/+/node/+/data"

# Optional: Customize API port
export API_PORT="8080"

# Optional: Customize Redis settings (for rate limiting)
export REDIS_URL="localhost:6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"
```

### 3. **Run the Application**
   ```bash
   go run cmd/main.go
   ```

### 4. **Verify Operation**
- Check console output for 60-second averages
- Visit `http://localhost:8080/health` for overall system health
- Visit `http://localhost:8080/health/database` for database status
- Visit `http://localhost:8080/health/mqtt` for MQTT status
- Visit `http://localhost:8080/metrics` for Prometheus metrics

## 🌐 REST API Endpoints

### **Health Checks**

#### Overall System Health
```bash
GET /health
```

#### Database Health
```bash
GET /health/database
```

#### MQTT Health
```bash
GET /health/mqtt
```

### **Sensor Data**

#### Current In-Memory Averages (All Nodes)
```bash
GET /sensors/averages
GET /sensors/averages?sensors=S1,S2,S3
GET /sensors/averages?greenhouse_id=GH1&node_id=Node01
```
- Returns the running average for the current 60-second window (not yet written to DB).

#### Latest Stored Averages (All Nodes, from Database)
```bash
GET /sensors/averages/latest
GET /sensors/averages/latest?node_id=Node03&sensors=S5
GET /sensors/averages/latest?greenhouse_id=GH1&node_id=Node03&sensors=S5,S7
```
- Returns the most recently stored average for each node from InfluxDB.
- Supports filtering by greenhouse_id, node_id, and sensors.

#### All Historical Averages (All Nodes, from Database)
```bash
GET /sensors/averages/all
GET /sensors/averages/all?node_id=Node03&sensors=S5
```
- Returns all historical averages for all nodes from InfluxDB.
- Supports filtering by greenhouse_id, node_id, and sensors.

**Sample Response:**
```json
[
  {
    "greenhouse_id": "GH1",
    "node_id": "Node03",
    "sensors": {
      "S5": 45.77,
      "S1": 12.34,
      ...
    }
  },
  ...
]
```

### **Monitoring**

#### Prometheus Metrics
```bash
GET /metrics
```
Returns Prometheus-formatted metrics for monitoring.

## 📊 Console Output

The application provides clean, informative console output:
- Only logs 60-second averages when they are uploaded to the database.
- No per-message logging for incoming MQTT data.

## ⚙️ Configuration

### **Environment Variables**

| Variable | Default | Description |
|----------|---------|-------------|
| `MQTT_BROKER` | `192.168.20.1` | MQTT broker IP address |
| `MQTT_PORT` | `1883` | MQTT broker port |
| `MQTT_TOPIC` | `greenhouse/+/node/+/data` | MQTT topic to subscribe to |
| `MQTT_CLIENT_ID` | `go-mqtt-subscriber-{timestamp}` | MQTT client identifier |
| `MQTT_USERNAME` | `` | MQTT username (optional) |
| `MQTT_PASSWORD` | `` | MQTT password (optional) |
| `INFLUXDB_URL` | `http://localhost:8086` | InfluxDB server URL |
| `INFLUXDB_TOKEN` | `[hardcoded]` | InfluxDB authentication token |
| `INFLUXDB_ORG` | `iot-agriculture` | InfluxDB organization |
| `INFLUXDB_BUCKET` | `sensor_data` | InfluxDB bucket for sensor data |
| `API_PORT` | `8080` | API server port |
| `REDIS_URL` | `localhost:6379` | Redis server URL for rate limiting |
| `REDIS_PASSWORD` | `` | Redis password (optional) |
| `REDIS_DB` | `0` | Redis database number |

### **ESP32 Data Format**

The backend expects JSON data from up to 5 ESP32 nodes, each publishing to topics of the form:

```
greenhouse/{greenhouse_id}/node/{node_id}/data
```

Example topics:
- `greenhouse/GH1/node/Node01/data`
- `greenhouse/GH1/node/Node02/data`
- ... up to Node05

Payload example:
```json
{
  "greenhouse_id": "GH1",
  "node_id": "Node01",
  "timestamp": 12345678, // optional, milliseconds since boot
  "S1": 12,
  "S2": 34,
  "S3": 56,
  "S4": 78,
  "S5": 90,
  "S6": 23,
  "S7": 45,
  "S8": 67,
  "S9": 89
}
```

## 📈 Monitoring & Metrics

### **Available Metrics**

#### MQTT Metrics
- `mqtt_messages_received_total` - Total MQTT messages received
- `mqtt_connection_status` - Connection status (0/1)
- `mqtt_reconnection_count_total` - Reconnection attempts

#### Sensor Metrics
- `sensor_readings_processed_total` - Total sensor readings processed
- `sensor_averages_calculated_total` - Total averages calculated

#### Database Metrics
- `influxdb_writes_total` - Successful InfluxDB writes
- `influxdb_write_errors_total` - InfluxDB write errors
- `influxdb_connection_status` - Connection status (0/1)

#### API Metrics
- `api_requests_total` - Request counts by method/endpoint/status
- `api_request_duration_seconds` - Response times

#### System Metrics
- `application_uptime_seconds` - Application uptime

### **Prometheus Configuration**

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'iot-agriculture-backend'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### **Grafana Dashboard**

Key metrics to monitor:
- **MQTT Connection Status**: `mqtt_connection_status`
- **Sensor Readings Rate**: `rate(sensor_readings_processed_total[5m])`
- **API Request Rate**: `rate(api_requests_total[5m])`
- **InfluxDB Write Errors**: `rate(influxdb_write_errors_total[5m])`
- **Application Uptime**: `application_uptime_seconds`

## 🚀 Production Deployment

### **Building for Production**
```bash
go build -o bin/iot-backend cmd/main.go
```

### **Running in Production**
```bash
# Set environment variables
export INFLUXDB_TOKEN="your-secure-token-here"

# Run the application
./bin/iot-backend
```

### **Docker Support**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### **Health Monitoring**
Monitor these endpoints for system health:
- `GET /health` - Overall system health (load balancer endpoint)
- `GET /health/database` - Database connectivity
- `GET /health/mqtt` - MQTT connectivity  
- `GET /metrics` - Prometheus metrics

## 🔧 Troubleshooting

### **Common Issues**

1. **InfluxDB Connection Issues**
   - Verify InfluxDB is running on localhost:8086
   - Check token permissions and organization/bucket access
   - Review `/health/database` endpoint

2. **MQTT Connection Issues**
   - Verify MQTT broker is running and accessible
   - Check network connectivity to broker
   - Review `/health/mqtt` endpoint

3. **Redis Connection Issues**
   - Verify Redis is running on localhost:6379
   - Check Redis authentication if configured
   - Rate limiting will be disabled if Redis is unavailable

4. **No Sensor Data**
   - Verify ESP32 is publishing to correct topic
   - Check MQTT subscription status
   - Verify JSON format matches expected structure

5. **API Returns 0 for Sensor Values**
   - Ensure InfluxDB field names are `s1_average`, ..., `s9_average` (lowercase, with `_average`)
   - The API maps these to `S1`, ..., `S9` in the response
   - If you see 0, check that data is being written for that node/sensor

6. **High Memory Usage**
   - Monitor memory metrics in Prometheus
   - Check for memory leaks in long-running deployments

7. **Rate Limiting Issues**
   - Check Redis connectivity
   - Verify rate limit headers in API responses
   - Monitor rate limit metrics in Prometheus

### **Log Analysis**
- **Clean console output** makes it easy to spot issues
- **Structured logging** provides consistent format
- **Health endpoints** provide quick status checks
- **Prometheus metrics** enable detailed monitoring

## 📝 License

This project is licensed under the MIT License. 

---

**Built with ❤️ for IoT Agriculture Systems** 