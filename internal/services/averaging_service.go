package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"iot-agriculture-backend/internal/models"
)

// AveragingService handles sensor data averaging calculations
type AveragingService struct {
	mu      sync.Mutex
	buffers map[string]*models.SensorAverages // key: greenhouse_id|node_id
}

// NewAveragingService creates a new averaging service
func NewAveragingService() *AveragingService {
	return &AveragingService{
		buffers: make(map[string]*models.SensorAverages),
	}
}

// AddSensorData adds sensor data to the averaging system
func (a *AveragingService) AddSensorData(data models.ESP32SensorData) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := data.GreenhouseID + "|" + data.NodeID
	buf, ok := a.buffers[key]
	if !ok {
		buf = &models.SensorAverages{
			GreenhouseID: data.GreenhouseID,
			NodeID:       data.NodeID,
			StartTime:    time.Now(),
		}
		a.buffers[key] = buf
	}
	if data.BagTemp != nil {
		buf.BagTemp = append(buf.BagTemp, *data.BagTemp)
	}
	if data.LightPar != nil {
		buf.LightPar = append(buf.LightPar, *data.LightPar)
	}
	if data.AirTemp != nil {
		buf.AirTemp = append(buf.AirTemp, *data.AirTemp)
	}
	if data.AirRh != nil {
		buf.AirRh = append(buf.AirRh, *data.AirRh)
	}
	if data.LeafTemp != nil {
		buf.LeafTemp = append(buf.LeafTemp, *data.LeafTemp)
	}
	if data.DripWeight != nil {
		buf.DripWeight = append(buf.DripWeight, *data.DripWeight)
	}
	if data.BagRh1 != nil {
		buf.BagRh1 = append(buf.BagRh1, *data.BagRh1)
	}
	if data.BagRh2 != nil {
		buf.BagRh2 = append(buf.BagRh2, *data.BagRh2)
	}
	if data.BagRh3 != nil {
		buf.BagRh3 = append(buf.BagRh3, *data.BagRh3)
	}
	if data.BagRh4 != nil {
		buf.BagRh4 = append(buf.BagRh4, *data.BagRh4)
	}
	if data.Rain != nil {
		buf.Rain = append(buf.Rain, *data.Rain)
	}
}

// CalculateAndDisplayAverages calculates and displays 60-second averages for all nodes
func (a *AveragingService) CalculateAndDisplayAverages() {
	a.CalculateAndDisplayAveragesWithLogging(nil, nil)
}

// CalculateAndDisplayAveragesWithLogging calculates, displays, and logs 60-second averages for all nodes
func (a *AveragingService) CalculateAndDisplayAveragesWithLogging(influxService *InfluxDBService, metricsService *MetricsService) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.buffers) == 0 {
		fmt.Printf("No sensor data to average in this period.\n")
		return
	}
	for _, buf := range a.buffers {
		result := calculateAveragesForBuffer(buf)
		displayAveragesForResult(result)
		if influxService != nil && influxService.IsConnected() && result.Readings > 0 {
			if err := influxService.LogAverages(result); err != nil {
				fmt.Printf("Warning: Failed to log to InfluxDB: %v\n", err)
				if metricsService != nil {
					metricsService.IncrementInfluxDBWriteErrors()
				}
			} else {
				if metricsService != nil {
					metricsService.IncrementInfluxDBWrites()
				}
			}
		} else if result.Readings == 0 {
			fmt.Printf("Skipping InfluxDB log - no sensor readings for %s/%s in this period\n", buf.GreenhouseID, buf.NodeID)
		}
	}
	// Clear all buffers for next period
	a.buffers = make(map[string]*models.SensorAverages)
}

// GetAverages returns the current averages for all nodes
func (a *AveragingService) GetAverages() []models.AverageResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	results := make([]models.AverageResult, 0, len(a.buffers))
	for _, buf := range a.buffers {
		results = append(results, calculateAveragesForBuffer(buf))
	}
	return results
}

// calculateAveragesForBuffer calculates the averages for a single node buffer
func calculateAveragesForBuffer(buf *models.SensorAverages) models.AverageResult {
	duration := time.Since(buf.StartTime)
	result := models.AverageResult{
		GreenhouseID: buf.GreenhouseID,
		NodeID:       buf.NodeID,
		Duration:     duration.Seconds(),
		Readings:     0,
	}
	if len(buf.BagTemp) > 0 {
		avg := calculateAverage(buf.BagTemp)
		result.BagTemp = &avg
		result.Readings = len(buf.BagTemp)
	}
	if len(buf.LightPar) > 0 {
		avg := calculateAverage(buf.LightPar)
		result.LightPar = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.LightPar)
		}
	}
	if len(buf.AirTemp) > 0 {
		avg := calculateAverage(buf.AirTemp)
		result.AirTemp = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.AirTemp)
		}
	}
	if len(buf.AirRh) > 0 {
		avg := calculateAverage(buf.AirRh)
		result.AirRh = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.AirRh)
		}
	}
	if len(buf.LeafTemp) > 0 {
		avg := calculateAverage(buf.LeafTemp)
		result.LeafTemp = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.LeafTemp)
		}
	}
	if len(buf.DripWeight) > 0 {
		avg := calculateAverage(buf.DripWeight)
		result.DripWeight = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.DripWeight)
		}
	}
	if len(buf.BagRh1) > 0 {
		avg := calculateAverage(buf.BagRh1)
		result.BagRh1 = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.BagRh1)
		}
	}
	if len(buf.BagRh2) > 0 {
		avg := calculateAverage(buf.BagRh2)
		result.BagRh2 = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.BagRh2)
		}
	}
	if len(buf.BagRh3) > 0 {
		avg := calculateAverage(buf.BagRh3)
		result.BagRh3 = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.BagRh3)
		}
	}
	if len(buf.BagRh4) > 0 {
		avg := calculateAverage(buf.BagRh4)
		result.BagRh4 = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.BagRh4)
		}
	}
	if len(buf.Rain) > 0 {
		avg := calculateAverage(buf.Rain)
		result.Rain = &avg
		if result.Readings == 0 {
			result.Readings = len(buf.Rain)
		}
	}
	return result
}

// displayAveragesForResult displays the calculated averages for a single node
func displayAveragesForResult(result models.AverageResult) {
	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Println("🕐 60-SECOND SENSOR AVERAGES")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏱️  Duration: %.1f seconds\n", result.Duration)
	fmt.Printf("🏠  Greenhouse: %s\n", result.GreenhouseID)
	fmt.Printf("📡  Node: %s\n", result.NodeID)
	fmt.Printf("📊  Total Readings: %d\n", result.Readings)
	fmt.Println(strings.Repeat("-", 60))
	if result.BagTemp != nil {
		fmt.Printf("🌡️  Bag_Temp: %.2f\n", *result.BagTemp)
	}
	if result.LightPar != nil {
		fmt.Printf("💡 Light_Par: %.2f\n", *result.LightPar)
	}
	if result.AirTemp != nil {
		fmt.Printf("🌡️  Air_Temp: %.2f\n", *result.AirTemp)
	}
	if result.AirRh != nil {
		fmt.Printf("💧 Air_Rh: %.2f\n", *result.AirRh)
	}
	if result.LeafTemp != nil {
		fmt.Printf("🌿 Leaf_temp: %.2f\n", *result.LeafTemp)
	}
	if result.DripWeight != nil {
		fmt.Printf("⚖️  drip_weight: %.2f\n", *result.DripWeight)
	}
	if result.BagRh1 != nil {
		fmt.Printf("💧 Bag_Rh1: %.2f\n", *result.BagRh1)
	}
	if result.BagRh2 != nil {
		fmt.Printf("💧 Bag_Rh2: %.2f\n", *result.BagRh2)
	}
	if result.BagRh3 != nil {
		fmt.Printf("💧 Bag_Rh3: %.2f\n", *result.BagRh3)
	}
	if result.BagRh4 != nil {
		fmt.Printf("💧 Bag_Rh4: %.2f\n", *result.BagRh4)
	}
	if result.Rain != nil {
		fmt.Printf("🌧️  Rain: %.2f\n", *result.Rain)
	}
	fmt.Println(strings.Repeat("=", 60) + "\n")

	if result.Readings == 0 {
		fmt.Println("⚠️  WARNING: No sensor readings received in the last 60 seconds for this node!")
		fmt.Println("   Check if ESP32 is sending data to topic for this node")
		fmt.Println("   Check MQTT broker connectivity")
		fmt.Println("   This period will NOT be logged to InfluxDB")
	}
}

// (resetAverages is no longer needed; buffers are cleared in CalculateAndDisplayAveragesWithLogging)

// GetReadingCount returns the current number of readings
func (a *AveragingService) GetReadingCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	count := 0
	for _, buf := range a.buffers {
		if len(buf.BagTemp) > 0 {
			count += len(buf.BagTemp)
		}
		if len(buf.LightPar) > 0 {
			count += len(buf.LightPar)
		}
		if len(buf.AirTemp) > 0 {
			count += len(buf.AirTemp)
		}
		if len(buf.AirRh) > 0 {
			count += len(buf.AirRh)
		}
		if len(buf.LeafTemp) > 0 {
			count += len(buf.LeafTemp)
		}
		if len(buf.DripWeight) > 0 {
			count += len(buf.DripWeight)
		}
		if len(buf.BagRh1) > 0 {
			count += len(buf.BagRh1)
		}
		if len(buf.BagRh2) > 0 {
			count += len(buf.BagRh2)
		}
		if len(buf.BagRh3) > 0 {
			count += len(buf.BagRh3)
		}
		if len(buf.BagRh4) > 0 {
			count += len(buf.BagRh4)
		}
		if len(buf.Rain) > 0 {
			count += len(buf.Rain)
		}
	}
	return count
}

// GetDuration returns the current duration since last reset
func (a *AveragingService) GetDuration() time.Duration {
	// Not meaningful for multi-node; return 0
	return 0
}

// calculateAverage calculates the average of a slice of integers
func calculateAverage(values []int) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0
	for _, v := range values {
		sum += v
	}

	result := float64(sum) / float64(len(values))

	return result
}
