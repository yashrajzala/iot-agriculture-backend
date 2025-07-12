# IoT Agriculture Backend

A lean, production-ready Go backend for IoT agriculture systems that processes sensor data from ESP32 devices via MQTT, calculates 60-second averages, logs data to InfluxDB, and provides REST APIs for frontend integration.

## 🎯 Features

- **MQTT Sensor Data Processing**: Real-time processing of ESP32 sensor data
- **60-Second Averaging**: Automatic calculation and logging of sensor averages
- **InfluxDB Integration**: Efficient time-series data storage
- **REST API Endpoints**: Health checks and sensor data retrieval
- **Production Ready**: Clean, optimized code with proper error handling
- **Modular Architecture**: Clean separation of concerns

## 🏗️ Architecture

```
iot-agriculture-backend/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── api/                       # REST API endpoints
│   │   ├── api.go                 # Main API server setup
│   │   ├── middleware.go          # CORS and common middleware
│   │   ├── database_health.go     # Database health check API
│   │   ├── mqtt_health.go         # MQTT connection health API
│   │   ├── sensor_averages.go     # Sensor averages data API
│   │   └── README.md              # API documentation
│   ├── config/                    # Configuration management
│   ├── models/                    # Data models (ESP32 sensor data)
│   ├── mqtt/                      # MQTT client abstraction
│   └── services/                  # Business logic services
│       ├── sensor_service.go      # Sensor data processing
│       ├── averaging_service.go   # 60-second averaging logic
│       └── influxdb_service.go    # InfluxDB integration
├── go.mod                         # Go module dependencies
└── README.md                      # This file
```

## 📋 Prerequisites

- **Go 1.24.5** or higher
- **MQTT Broker** (e.g., Mosquitto, HiveMQ, AWS IoT)
- **InfluxDB 2.x** running on localhost:8086
- **ESP32 Device** publishing sensor data

## 🚀 Quick Start

1. **Clone and navigate to the project**
   ```bash
   git clone <repository-url>
   cd iot-agriculture-backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure environment variables** (optional - uses defaults)
   ```bash
   # Set environment variables or use defaults:
   # MQTT_BROKER=192.168.20.1
   # MQTT_PORT=1883
   # MQTT_TOPIC=esp32/data
   # MQTT_CLIENT_ID=go-mqtt-subscriber
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```

5. **Stop the application**
   ```bash
   # Press Ctrl+C for graceful shutdown
   ```

## 🌐 REST API Endpoints

The backend provides REST APIs for frontend integration. All endpoints run on-demand and support CORS.

### Available Endpoints

#### 1. Database Health Check
- **Endpoint:** `GET /health/database`
- **Description:** Checks InfluxDB connection status
- **Response:**
```json
{
  "status": "connected",
  "connected": true,
  "message": "Connected to InfluxDB - Org: iot-agriculture, Bucket: sensor_data",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### 2. MQTT Connection Health Check
- **Endpoint:** `GET /health/mqtt`
- **Description:** Checks MQTT broker connection status
- **Response:**
```json
{
  "status": "connected",
  "connected": true,
  "message": "Connected to MQTT broker: tcp://192.168.20.1:1883",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### 3. Sensor Averages
- **Endpoint:** `GET /sensors/averages`
- **Description:** Fetches current sensor averages
- **Query Parameters:**
  - `sensors` (optional): Comma-separated list (e.g., "S1,S2,S3" or "all")
  - `greenhouse_id` (optional): Filter by greenhouse
  - `node_id` (optional): Filter by node
- **Response:**
```json
{
  "greenhouse_id": "GH1",
  "node_id": "Node01",
  "duration": 60.5,
  "readings": 15,
  "timestamp": "2024-01-01T12:00:00Z",
  "sensors": {
    "S1": 25.5,
    "S2": 30.2,
    "S3": 28.7,
    "S4": 32.1,
    "S5": 29.8,
    "S6": 27.3,
    "S7": 31.4,
    "S8": 26.9,
    "S9": 33.6
  }
}
```

### API Usage Examples

```bash
# Get all sensor averages
curl http://localhost:8080/sensors/averages

# Get specific sensors (S1, S5, S9)
curl "http://localhost:8080/sensors/averages?sensors=S1,S5,S9"

# Check database health
curl http://localhost:8080/health/database

# Check MQTT health
curl http://localhost:8080/health/mqtt
```

### Frontend Integration

```javascript
// Get S5 sensor average
fetch('http://localhost:8080/sensors/averages?sensors=S5')
  .then(response => response.json())
  .then(data => {
    console.log('S5 Average:', data.sensors.S5);
  });

// Check system health
fetch('http://localhost:8080/health/database')
  .then(response => response.json())
  .then(data => {
    console.log('Database Status:', data.status);
  });
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MQTT_BROKER` | `192.168.20.1` | MQTT broker IP address |
| `MQTT_PORT` | `1883` | MQTT broker port |
| `MQTT_TOPIC` | `esp32/data` | MQTT topic to subscribe to |
| `MQTT_CLIENT_ID` | `go-mqtt-subscriber` | MQTT client identifier |
| `MQTT_USERNAME` | `` | MQTT username (optional) |
| `MQTT_PASSWORD` | `` | MQTT password (optional) |

### ESP32 Data Format

The backend expects JSON data from ESP32 devices in this format:

```json
{
  "greenhouse_id": "GH1",
  "node_id": "Node01",
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

## 📊 Data Processing

### Real-time Processing
- **Individual Readings**: Each MQTT message is parsed and displayed
- **JSON Validation**: Automatic validation of incoming sensor data
- **Error Handling**: Graceful handling of malformed data

### 60-Second Averaging
- **Automatic Calculation**: Averages calculated every 60 seconds
- **Accumulative**: Collects all readings during the period
- **Memory Efficient**: Clears old data after averaging
- **InfluxDB Logging**: Logs averages to time-series database

### Sample Output
```
=== New ESP32 Sensor Data ===
Topic: esp32/data
Payload: {"greenhouse_id":"GH1","node_id":"Node01","S1":12,"S2":34,...}
Greenhouse: GH1
Node: Node01
S1: 12
S2: 34
...
============================

=== 60-Second Sensor Averages ===
Duration: 60.1 seconds
Greenhouse: GH1
Node: Node01
S1 Average: 23.45 (from 12 readings)
S2 Average: 67.89 (from 12 readings)
...
================================
```

## 🏛️ Services Architecture

### SensorService (`internal/services/sensor_service.go`)
- **Responsibility**: MQTT data processing and display
- **Functions**: JSON parsing, data validation, real-time display
- **Integration**: Delegates averaging to AveragingService

### AveragingService (`internal/services/averaging_service.go`)
- **Responsibility**: 60-second averaging calculations
- **Functions**: Data accumulation, average calculation, statistics
- **Features**: Reading count, duration tracking, automatic reset

### InfluxDBService (`internal/services/influxdb_service.go`)
- **Responsibility**: Time-series data logging
- **Functions**: Connection management, data writing, error handling
- **Features**: Automatic reconnection, efficient batch writing

## 🚀 Production Deployment

### Building for Production
```bash
go build -o bin/iot-backend cmd/main.go
```

### Running in Production
```bash
./bin/iot-backend
```

### Docker Support
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

## 🔧 Troubleshooting

### Common Issues

1. **MQTT Connection Issues**
   - Verify MQTT broker is running
   - Check network connectivity
   - Verify broker IP and port

2. **InfluxDB Connection Issues**
   - Ensure InfluxDB is running on localhost:8086
   - Verify token and organization settings
   - Check bucket exists

3. **No Sensor Data**
   - Verify ESP32 is publishing to correct topic
   - Check MQTT subscription
   - Verify JSON format matches expected structure

## 📝 License

This project is licensed under the MIT License. 