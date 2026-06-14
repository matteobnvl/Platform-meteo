package detector

import (
	"testing"
	"time"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

var defaultCfg = Config{
	StormWindThreshold:    80,
	StormDurationHours:    3,
	HeatwaveTempThreshold: 35,
	HeatwaveDurationDays:  3,
	ColdwaveTempThreshold: 0,
	ColdwaveDurationDays:  3,
	FloodPrecipThreshold:  50,
}

// test tempete

func TestDetectStorms_DetecteUneTemplete(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, WindSpeed: 90},
		{ObservedAt: base.Add(1 * time.Hour), WindSpeed: 95},
		{ObservedAt: base.Add(2 * time.Hour), WindSpeed: 100},
		{ObservedAt: base.Add(3 * time.Hour), WindSpeed: 85},
	}

	events := DetectStorms(obs, defaultCfg)

	if len(events) != 1 {
		t.Fatalf("attendu 1 événement storm, obtenu %d", len(events))
	}
	if events[0].Type != "storm" {
		t.Errorf("attendu type=storm, obtenu %s", events[0].Type)
	}
}

func TestDetectStorms_PasDeTemplete(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, WindSpeed: 50},
		{ObservedAt: base.Add(1 * time.Hour), WindSpeed: 60},
	}

	events := DetectStorms(obs, defaultCfg)

	if len(events) != 0 {
		t.Errorf("attendu 0 événement, obtenu %d", len(events))
	}
}

// test canicule

func TestDetectHeatwaves_DetecteUneCanicule(t *testing.T) {
	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, Temperature: 36},
		{ObservedAt: base.Add(24 * time.Hour), Temperature: 37},
		{ObservedAt: base.Add(48 * time.Hour), Temperature: 38},
	}

	events := DetectHeatwaves(obs, defaultCfg)

	if len(events) != 1 {
		t.Fatalf("attendu 1 événement heatwave, obtenu %d", len(events))
	}
	if events[0].Type != "heatwave" {
		t.Errorf("attendu type=heatwave, obtenu %s", events[0].Type)
	}
}

func TestDetectHeatwaves_PasDeCanicule(t *testing.T) {
	base := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, Temperature: 30},
		{ObservedAt: base.Add(24 * time.Hour), Temperature: 32},
	}

	events := DetectHeatwaves(obs, defaultCfg)

	if len(events) != 0 {
		t.Errorf("attendu 0 événement, obtenu %d", len(events))
	}
}

// test vague froid

func TestDetectColdwaves_DetecteUneVagueDeFroid(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, Temperature: -2},
		{ObservedAt: base.Add(24 * time.Hour), Temperature: -5},
		{ObservedAt: base.Add(48 * time.Hour), Temperature: -3},
	}

	events := DetectColdwaves(obs, defaultCfg)

	if len(events) != 1 {
		t.Fatalf("attendu 1 événement cold_wave, obtenu %d", len(events))
	}
	if events[0].Type != "cold_wave" {
		t.Errorf("attendu type=cold_wave, obtenu %s", events[0].Type)
	}
}

// test inondation

func TestDetectFloods_DetecteUneInondation(t *testing.T) {
	base := time.Date(2026, 3, 1, 6, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, Precipitation: 30},
		{ObservedAt: base.Add(6 * time.Hour), Precipitation: 25},
	}

	events := DetectFloods(obs, defaultCfg)

	if len(events) != 1 {
		t.Fatalf("attendu 1 événement flood, obtenu %d", len(events))
	}
	if events[0].Type != "flood" {
		t.Errorf("attendu type=flood, obtenu %s", events[0].Type)
	}
}

func TestDetectFloods_PasInondation(t *testing.T) {
	base := time.Date(2026, 3, 1, 6, 0, 0, 0, time.UTC)
	obs := []model.Observation{
		{ObservedAt: base, Precipitation: 10},
		{ObservedAt: base.Add(6 * time.Hour), Precipitation: 5},
	}

	events := DetectFloods(obs, defaultCfg)

	if len(events) != 0 {
		t.Errorf("attendu 0 événement, obtenu %d", len(events))
	}
}
