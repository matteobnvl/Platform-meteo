package mapping

import (
	"testing"
)

func TestTempKelvinToCelsius(t *testing.T) {
	obs := synopObservation{Temperature: 273.15}
	got := obs.getTempToCelsius()
	if got != 0.0 {
		t.Errorf("attendu 0.0°C, obtenu %f", got)
	}
}

func TestTempKelvinToCelsius_Boiling(t *testing.T) {
	obs := synopObservation{Temperature: 373.15}
	got := obs.getTempToCelsius()
	if got != 100.0 {
		t.Errorf("attendu 100.0°C, obtenu %f", got)
	}
}

func TestWindSpeedMsToKmh(t *testing.T) {
	obs := synopObservation{WindSpeed: 10}
	got := obs.getWindSpeedToKmh()
	if got != 36.0 {
		t.Errorf("attendu 36.0 km/h, obtenu %f", got)
	}
}

func TestMapSynopStation(t *testing.T) {
	header := []string{"geo_id_wmo", "name", "lat", "lon"}
	row := []string{"07149", "Paris-Montsouris", "48.8", "2.3"}

	st := MapSynopStation(header, row)

	if st.Id != "07149" {
		t.Errorf("attendu id=07149, obtenu %s", st.Id)
	}
	if st.Name != "Paris-Montsouris" {
		t.Errorf("attendu name=Paris-Montsouris, obtenu %s", st.Name)
	}
	if st.Latitude != 48.8 {
		t.Errorf("attendu lat=48.8, obtenu %f", st.Latitude)
	}
}

func TestMapSynopObservation_ConvertsUnits(t *testing.T) {
	header := []string{"geo_id_wmo", "validity_time", "t", "ff", "dd", "rr1"}
	// 293.15 K = 20°C - 5 m/s = 18 km/h
	row := []string{"07149", "2026-01-01T12:00:00Z", "293.15", "5", "180", "2.5"}

	obs := MapSynopObservation(header, row)

	wantTemp := 20.0
	if obs.Temperature != wantTemp {
		t.Errorf("température : attendu %f°C, obtenu %f", wantTemp, obs.Temperature)
	}

	wantWind := 18.0
	if obs.WindSpeed != wantWind {
		t.Errorf("vent : attendu %f km/h, obtenu %f", wantWind, obs.WindSpeed)
	}

	if obs.Precipitation != 2.5 {
		t.Errorf("précipitations : attendu 2.5, obtenu %f", obs.Precipitation)
	}
}
