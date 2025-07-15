package models

import "time"

// ESP32SensorData matches the JSON published by the ESP32
// Only keep fields needed for MQTT sensor data
// Example: {"greenhouse_id":"GH1","node_id":"Node01","S1":12,...,"S9":85}
type ESP32SensorData struct {
	GreenhouseID string `json:"greenhouse_id"`
	NodeID       string `json:"node_id"`
	Timestamp    *int64 `json:"timestamp,omitempty"` // Optional timestamp from ESP32
	S1           int    `json:"S1"`
	S2           int    `json:"S2"`
	S3           int    `json:"S3"`
	S4           int    `json:"S4"`
	S5           int    `json:"S5"`
	S6           int    `json:"S6"`
	S7           int    `json:"S7"`
	S8           int    `json:"S8"`
	S9           int    `json:"S9"`
}

// SensorAverages holds the accumulated values for averaging
type SensorAverages struct {
	GreenhouseID string
	NodeID       string
	S1Values     []int
	S2Values     []int
	S3Values     []int
	S4Values     []int
	S5Values     []int
	S6Values     []int
	S7Values     []int
	S8Values     []int
	S9Values     []int
	StartTime    time.Time
}

// AverageResult represents the calculated averages
type AverageResult struct {
	GreenhouseID string
	NodeID       string
	Duration     float64
	Readings     int
	S1Average    float64
	S2Average    float64
	S3Average    float64
	S4Average    float64
	S5Average    float64
	S6Average    float64
	S7Average    float64
	S8Average    float64
	S9Average    float64
}
