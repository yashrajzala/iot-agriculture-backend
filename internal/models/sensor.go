package models

import "time"

// ESP32SensorData matches the JSON published by the ESP32
// Updated for new sensor names from Arduino/ESP32
// Node01-04: Bag_Temp, Light_Par, Air_Temp, Air_Rh, Leaf_temp, drip_weight, Bag_Rh1, Bag_Rh2, Bag_Rh3, Bag_Rh4
// Node05: Light_Par, Air_Temp, Air_Rh, Rain
// All fields are optional to support different node payloads
// Example: {"greenhouse_id":"GH1","node_id":"Node01","Bag_Temp":12,...}
type ESP32SensorData struct {
	GreenhouseID string `json:"greenhouse_id"`
	NodeID       string `json:"node_id"`
	Timestamp    *int64 `json:"timestamp,omitempty"`
	BagTemp      *int   `json:"Bag_Temp,omitempty"`
	LightPar     *int   `json:"Light_Par,omitempty"`
	AirTemp      *int   `json:"Air_Temp,omitempty"`
	AirRh        *int   `json:"Air_Rh,omitempty"`
	LeafTemp     *int   `json:"Leaf_temp,omitempty"`
	DripWeight   *int   `json:"drip_weight,omitempty"`
	BagRh1       *int   `json:"Bag_Rh1,omitempty"`
	BagRh2       *int   `json:"Bag_Rh2,omitempty"`
	BagRh3       *int   `json:"Bag_Rh3,omitempty"`
	BagRh4       *int   `json:"Bag_Rh4,omitempty"`
	Rain         *int   `json:"Rain,omitempty"`
}

// SensorAverages holds the accumulated values for averaging (all fields optional)
type SensorAverages struct {
	GreenhouseID string
	NodeID       string
	BagTemp      []int
	LightPar     []int
	AirTemp      []int
	AirRh        []int
	LeafTemp     []int
	DripWeight   []int
	BagRh1       []int
	BagRh2       []int
	BagRh3       []int
	BagRh4       []int
	Rain         []int
	StartTime    time.Time
}

// AverageResult represents the calculated averages (all fields optional)
type AverageResult struct {
	GreenhouseID string
	NodeID       string
	Duration     float64
	Readings     int
	BagTemp      *float64
	LightPar     *float64
	AirTemp      *float64
	AirRh        *float64
	LeafTemp     *float64
	DripWeight   *float64
	BagRh1       *float64
	BagRh2       *float64
	BagRh3       *float64
	BagRh4       *float64
	Rain         *float64
}
