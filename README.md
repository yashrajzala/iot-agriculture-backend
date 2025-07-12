# IoT Agriculture Backend

A lean, production-ready Go backend for IoT agriculture systems that processes sensor data from ESP32 devices via MQTT, calculates 60-second averages, and logs data to InfluxDB.

## 🎯 Features

- **MQTT Sensor Data Processing**: Real-time processing of ESP32 sensor data
- **60-Second Averaging**: Automatic calculation and logging of sensor averages
- **InfluxDB Integration**: Efficient time-series data storage
- **Production Ready**: Clean, optimized code with proper error handling
- **Modular Architecture**: Clean separation of concerns

## 🏗️ Architecture

```
iot-agriculture-backend/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
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