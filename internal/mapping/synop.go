package mapping

import (
	"strconv"
	"time"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

type synopObservation struct {
	StationId      string
	Date           string
	Temperature    float64 // kelvin
	WindSpeed      float64 // m/s
	WindDirection  float64
	Precipitation1 float64 // mm
}

type synopStation struct {
	Id        string
	Name      string
	Latitude  float64
	Longitude float64
}

func MapSynopObservation(header []string, row []string) *model.Observation {
	cols := buildMap(header, row)
	s := parseSynopObservation(cols)

	return &model.Observation{
		Id:            0,
		StationId:     s.StationId,
		ObservedAt:    s.getDate(),
		Temperature:   s.getTempToCelsius(),
		WindSpeed:     s.getWindSpeedToKmh(),
		WindDirection: s.WindDirection,
		Precipitation: s.Precipitation1,
	}
}

func MapSynopStation(header []string, row []string) *model.Station {
	cols := buildMap(header, row)
	s := parseSynopStation(cols)

	return &model.Station{
		Id:        s.Id,
		Name:      s.Name,
		Latitude:  s.Latitude,
		Longitude: s.Longitude,
	}
}

func parseSynopObservation(obs map[string]string) synopObservation {
	temp, _ := strconv.ParseFloat(obs["t"], 64)
	ff, _ := strconv.ParseFloat(obs["ff"], 64)
	dd, _ := strconv.ParseFloat(obs["dd"], 64)
	rr1, _ := strconv.ParseFloat(obs["rr1"], 64)
	return synopObservation{
		StationId:      obs["geo_id_wmo"],
		Date:           obs["validity_time"],
		Temperature:    temp,
		WindSpeed:      ff,
		WindDirection:  dd,
		Precipitation1: rr1,
	}
}

func parseSynopStation(st map[string]string) synopStation {
	lat, _ := strconv.ParseFloat(st["lat"], 64)
	lon, _ := strconv.ParseFloat(st["lon"], 64)
	return synopStation{
		Id:        st["geo_id_wmo"],
		Name:      st["name"],
		Latitude:  lat,
		Longitude: lon,
	}
}

func buildMap(header []string, row []string) map[string]string {
	cols := make(map[string]string, len(header))
	for i, name := range header {
		if i < len(row) {
			cols[name] = row[i]
		}
	}
	return cols
}

func (s synopObservation) getTempToCelsius() float64 {
	return s.Temperature - 273.15 // kelvin to celsius
}

func (s synopObservation) getWindSpeedToKmh() float64 {
	return s.WindSpeed * 3.6 // m/s to km/h
}

func (s synopObservation) getDate() time.Time {
	date, _ := time.Parse(time.RFC3339, s.Date)
	return date
}
