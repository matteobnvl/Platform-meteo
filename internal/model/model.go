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
	Id              int
	StationId       string
	ObservedAt      time.Time
	Temperature     *float64
	WindSpeed       *float64
	WindDirection   *float64
	Precipitation1  *float64
	Precipitation3  *float64
	Precipitation6  *float64
	Precipitation12 *float64
	Precipitation24 *float64
}

type Event struct {
	Id        int
	StationId string
	Type      string
	StartedAt time.Time
	EndedAt   *time.Time
	Metadata  map[string]any
}
