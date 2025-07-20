# API Documentation

This directory contains the API endpoints for the IoT Agriculture Backend. Each API endpoint is organized in its own file for better maintainability.

## API Structure

- `api.go` - Main API server setup and routing
- `middleware.go` - Common middleware functions (CORS, etc.)
- `database_health.go` - Database (InfluxDB) health check endpoint
- `mqtt_health.go` - MQTT connection health check endpoint
- `sensor_averages.go` - Sensor averages data endpoint

## Available Endpoints

### 1. Database Health Check
- **Endpoint:** `GET /health/database`
- **Description:** Checks the connection status of the local InfluxDB database
- **Response:**
```json
{
  "status": "connected|disconnected|unknown",
  "connected": true|false,
  "message": "Connection details or error message",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 2. MQTT Connection Health Check
- **Endpoint:** `GET /health/mqtt`
- **Description:** Checks the connection status of the MQTT broker
- **Response:**
```json
{
  "status": "connected|disconnected|unknown",
  "connected": true|false,
  "message": "Connection details or error message",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 3. Sensor Averages
- **Endpoint:** `GET /sensors/averages`
- **Description:** Fetches current sensor averages from the local database
- **Query Parameters:**
  - `sensors` (optional): Comma-separated list of sensors (e.g., "Bag_Temp,Light_Par,Air_Temp" or "all")
  - `greenhouse_id` (optional): Filter by specific greenhouse
  - `node_id` (optional): Filter by specific node
- **Response:**
```json
{
  "greenhouse_id": "GH1",
  "node_id": "Node01",
  "duration": 60.5,
  "readings": 15,
  "timestamp": "2024-01-01T12:00:00Z",
  "sensors": {
    "Bag_Temp": 25.5,
    "Light_Par": 30.2,
    "Air_Temp": 28.7,
    "Air_Rh": 32.1,
    "Leaf_temp": 29.8,
    "drip_weight": 27.3,
    "Bag_Rh1": 31.4,
    "Bag_Rh2": 26.9,
    "Bag_Rh3": 33.6
  }
}
```
- **Examples:**
  - `GET /sensors/averages` - Get all sensor averages
  - `GET /sensors/averages?sensors=Bag_Temp,Light_Par,Air_Temp` - Get only Bag_Temp, Light_Par, Air_Temp averages
  - `GET /sensors/averages?greenhouse_id=GH1&node_id=Node01` - Get averages for specific location

## Features

- **On-demand execution:** APIs only run when the frontend makes requests
- **CORS support:** All endpoints support cross-origin requests
- **Modular design:** Each endpoint is in its own file for easy maintenance
- **Consistent response format:** All health endpoints return the same JSON structure
- **Graceful shutdown:** API server shuts down properly with the main application

## Adding New Endpoints

To add a new API endpoint:

1. Create a new file (e.g., `new_endpoint.go`)
2. Define a handler struct with a `Handle` method
3. Register the route in `api.go` with CORS middleware
4. Update this documentation

Example:
```go
// new_endpoint.go
type NewEndpointHandler struct {
    // dependencies
}

func (h *NewEndpointHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // implementation
}

// api.go
newEndpointHandler := NewNewEndpointHandler(dependencies)
mux.HandleFunc("/new/endpoint", CORSMiddleware(newEndpointHandler.Handle))
```

## 🆕 Changelog

### vNext (Unreleased)
- API now supports real sensor names (Bag_Temp, Light_Par, etc.)
- Query params `node_id` and `sensors` allow filtering for specific node/sensor values
- All API responses and examples updated for new sensor format
- Asynchronous MQTT processing and batching for high throughput
- Context propagation through all service layers
- Resource pooling and concurrency tuning
- Final codebase review: modular, robust, production-ready 