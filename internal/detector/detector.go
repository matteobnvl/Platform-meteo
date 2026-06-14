package detector

import (
	"database/sql"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"time"

	dbpkg "github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/model"
)

type Config struct {
	StormWindThreshold    float64 // km/h
	StormDurationHours    int
	HeatwaveTempThreshold float64 // °C
	HeatwaveDurationDays  int
	ColdwaveTempThreshold float64 // °C
	ColdwaveDurationDays  int
	FloodPrecipThreshold  float64 // mm/24h
	IntervalMinutes       int
}

func LoadConfig() Config {
	return Config{
		StormWindThreshold:    parseFloat("STORM_WIND_THRESHOLD", 80),
		StormDurationHours:    parseInt("STORM_DURATION_HOURS", 3),
		HeatwaveTempThreshold: parseFloat("HEATWAVE_TEMP_THRESHOLD", 35),
		HeatwaveDurationDays:  parseInt("HEATWAVE_DURATION_DAYS", 3),
		ColdwaveTempThreshold: parseFloat("COLDWAVE_TEMP_THRESHOLD", 0),
		ColdwaveDurationDays:  parseInt("COLDWAVE_DURATION_DAYS", 3),
		FloodPrecipThreshold:  parseFloat("FLOOD_PRECIP_THRESHOLD", 50),
		IntervalMinutes:       parseInt("DETECTOR_INTERVAL_MINUTES", 60),
	}
}

func RunOnce(database *sql.DB, cfg Config) {
	stations, err := dbpkg.GetStations(database, "")
	if err != nil {
		slog.Error("détecteur : récupération stations", "err", err)
		return
	}

	results := make(chan int, len(stations))

	for _, st := range stations {
		go func(st model.Station) {
			results <- detectForStation(database, st, cfg)
		}(st)
	}

	total := 0
	for range stations {
		total += <-results
	}

	slog.Info("détection terminée", "nouveaux_événements", total)
}

func detectForStation(database *sql.DB, station model.Station, cfg Config) int {
	now := time.Now().UTC()
	since := now.AddDate(0, 0, -30)

	obs, err := dbpkg.GetObservationsByStation(database, station.Id, &since, &now, 0, 0)
	if err != nil || len(obs) == 0 {
		return 0
	}
	// GetObservationsByStation renvoie en DESC, on inverse pour avoir ASC
	for i, j := 0, len(obs)-1; i < j; i, j = i+1, j-1 {
		obs[i], obs[j] = obs[j], obs[i]
	}

	var detected []model.Event
	detected = append(detected, DetectStorms(obs, cfg)...)
	detected = append(detected, DetectHeatwaves(obs, cfg)...)
	detected = append(detected, DetectColdwaves(obs, cfg)...)
	detected = append(detected, DetectFloods(obs, cfg)...)

	count := 0
	for _, e := range detected {
		e.StationId = station.Id
		inserted, err := dbpkg.InsertEventIfNew(database, e)
		if err != nil {
			slog.Error("détecteur : insertion événement", "type", e.Type, "err", err)
			continue
		}
		if inserted {
			count++
		}
	}
	return count
}

// détecte les tempêtes : rafales > seuil pendant au moins N heures consécutives.
func DetectStorms(obs []model.Observation, cfg Config) []model.Event {
	var events []model.Event
	minDuration := time.Duration(cfg.StormDurationHours) * time.Hour

	i := 0
	for i < len(obs) {
		if obs[i].WindSpeed <= cfg.StormWindThreshold {
			i++
			continue
		}
		// début d'une séquence de vent fort
		j := i + 1
		for j < len(obs) &&
			obs[j].WindSpeed > cfg.StormWindThreshold &&
			obs[j].ObservedAt.Sub(obs[j-1].ObservedAt) <= 2*time.Hour {
			j++
		}
		duration := obs[j-1].ObservedAt.Sub(obs[i].ObservedAt)
		if duration >= minDuration {
			maxWind := 0.0
			for _, o := range obs[i:j] {
				if o.WindSpeed > maxWind {
					maxWind = o.WindSpeed
				}
			}
			endT := obs[j-1].ObservedAt
			events = append(events, model.Event{
				Type:      "storm",
				StartedAt: obs[i].ObservedAt,
				EndedAt:   &endT,
				Severity:  stormSeverity(maxWind),
				Metadata:  map[string]any{"max_wind_speed_kmh": maxWind, "duration_hours": duration.Hours()},
			})
		}
		i = j
	}
	return events
}

// détecte les canicules : température max >= seuil pendant N jours consécutifs.
func DetectHeatwaves(obs []model.Observation, cfg Config) []model.Event {
	dailyMax := buildDailyMax(obs)

	var events []model.Event
	i := 0
	for i < len(dailyMax) {
		if dailyMax[i].maxT < cfg.HeatwaveTempThreshold {
			i++
			continue
		}
		j := i + 1
		for j < len(dailyMax) &&
			dailyMax[j].maxT >= cfg.HeatwaveTempThreshold &&
			dailyMax[j].date.Sub(dailyMax[j-1].date) <= 25*time.Hour {
			j++
		}
		if j-i >= cfg.HeatwaveDurationDays {
			peak := dailyMax[i].maxT
			for _, d := range dailyMax[i:j] {
				if d.maxT > peak {
					peak = d.maxT
				}
			}
			endT := dailyMax[j-1].date.Add(24 * time.Hour)
			events = append(events, model.Event{
				Type:      "heatwave",
				StartedAt: dailyMax[i].date,
				EndedAt:   &endT,
				Severity:  heatwaveSeverity(peak),
				Metadata:  map[string]any{"duration_days": j - i, "peak_temperature": peak},
			})
		}
		i = j
	}
	return events
}

// détecte les vagues de froid : température max <= seuil pendant N jours consécutifs.
func DetectColdwaves(obs []model.Observation, cfg Config) []model.Event {
	dailyMax := buildDailyMax(obs)

	var events []model.Event
	i := 0
	for i < len(dailyMax) {
		if dailyMax[i].maxT > cfg.ColdwaveTempThreshold {
			i++
			continue
		}
		j := i + 1
		for j < len(dailyMax) &&
			dailyMax[j].maxT <= cfg.ColdwaveTempThreshold &&
			dailyMax[j].date.Sub(dailyMax[j-1].date) <= 25*time.Hour {
			j++
		}
		if j-i >= cfg.ColdwaveDurationDays {
			minT := dailyMax[i].maxT
			for _, d := range dailyMax[i:j] {
				if d.maxT < minT {
					minT = d.maxT
				}
			}
			endT := dailyMax[j-1].date.Add(24 * time.Hour)
			events = append(events, model.Event{
				Type:      "cold_wave",
				StartedAt: dailyMax[i].date,
				EndedAt:   &endT,
				Severity:  coldwaveSeverity(minT),
				Metadata:  map[string]any{"duration_days": j - i, "min_temperature": minT},
			})
		}
		i = j
	}
	return events
}

// détecte les inondations potentielles : précipitations > seuil en 24h.
func DetectFloods(obs []model.Observation, cfg Config) []model.Event {
	type dayAgg struct {
		total float64
		start time.Time
		end   time.Time
	}
	days := map[string]*dayAgg{}

	for _, o := range obs {
		key := o.ObservedAt.UTC().Format("2006-01-02")
		d, ok := days[key]
		if !ok {
			d = &dayAgg{start: o.ObservedAt, end: o.ObservedAt}
			days[key] = d
		}
		d.total += o.Precipitation
		if o.ObservedAt.Before(d.start) {
			d.start = o.ObservedAt
		}
		if o.ObservedAt.After(d.end) {
			d.end = o.ObservedAt
		}
	}

	var events []model.Event
	for day, d := range days {
		if d.total < cfg.FloodPrecipThreshold {
			continue
		}
		endT := d.end
		events = append(events, model.Event{
			Type:      "flood",
			StartedAt: d.start,
			EndedAt:   &endT,
			Severity:  floodSeverity(d.total),
			Metadata:  map[string]any{"total_precipitation_mm": d.total, "date": day},
		})
	}
	return events
}

type dayMaxTemp struct {
	date time.Time
	maxT float64
}

// regroupe les observations par jour et renvoie le max de température par jour, trié ASC.
func buildDailyMax(obs []model.Observation) []dayMaxTemp {
	maxByDay := map[string]float64{}
	for _, o := range obs {
		day := o.ObservedAt.UTC().Format("2006-01-02")
		if cur, ok := maxByDay[day]; !ok || o.Temperature > cur {
			maxByDay[day] = o.Temperature
		}
	}
	var days []dayMaxTemp
	for dayStr, maxT := range maxByDay {
		t, _ := time.Parse("2006-01-02", dayStr)
		days = append(days, dayMaxTemp{date: t, maxT: maxT})
	}
	sort.Slice(days, func(i, j int) bool { return days[i].date.Before(days[j].date) })
	return days
}

func stormSeverity(maxWind float64) string {
	if maxWind >= 150 {
		return "extreme"
	} else if maxWind >= 117 {
		return "high"
	}
	return "medium"
}

func heatwaveSeverity(maxT float64) string {
	if maxT >= 42 {
		return "extreme"
	} else if maxT >= 38 {
		return "high"
	}
	return "medium"
}

func coldwaveSeverity(minT float64) string {
	if minT <= -15 {
		return "extreme"
	} else if minT <= -10 {
		return "high"
	}
	return "medium"
}

func floodSeverity(totalMm float64) string {
	if totalMm >= 100 {
		return "extreme"
	} else if totalMm >= 75 {
		return "high"
	}
	return "medium"
}

func parseFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func parseInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
