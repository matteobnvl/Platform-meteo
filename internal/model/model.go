package model

import "time"

type Station struct {
	Id        string
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
}

type Observation struct {
	Id            int
	StationId     string
	ObservedAt    time.Time
	Temperature   float64
	WindSpeed     float64
	WindDirection float64
	Precipitation float64
}

type Event struct {
	Id        int
	StationId string
	Type      string
	StartedAt time.Time
	EndedAt   *time.Time
	Severity  string
	Metadata  map[string]any
}

type EventStat struct {
	Type    string
	Country string
	Count   int
}

type AggregateResult struct {
	Period  string  `json:"period"`
	AvgTemp float64 `json:"avg_temperature"`
	MaxTemp float64 `json:"max_temperature"`
	MinTemp float64 `json:"min_temperature"`
}
