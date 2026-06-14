package fetcher

import (
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/matteobnvl/Platform-meteo/internal/mapping"
	"github.com/matteobnvl/Platform-meteo/internal/model"
	"github.com/matteobnvl/Platform-meteo/internal/storage"
)

type jobData struct {
	obs     model.Observation
	station model.Station
}

const batchSize = 1000

func FetchSynop(db *sql.DB) {
	url := fmt.Sprintf(
		"https://object.files.data.gouv.fr/meteofrance/data/synchro_ftp/OBS/SYNOP/synop_%d.csv.gz",
		time.Now().Year(),
	)

	start := time.Now()
	slog.Info("ingestion démarrée", "url", url)

	resp, err := fetch(url)
	if err != nil {
		slog.Error("téléchargement échoué", "err", err)
		return
	}
	defer resp.Body.Close()

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		slog.Error("décompression gzip échouée", "err", err)
		return
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		slog.Error("lecture header CSV échouée", "err", err)
		return
	}

	knownStations := make(map[string]bool)
	var batch []model.Observation
	calls, errors := 0, 0

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Warn("ligne ignorée", "err", err)
			errors++
			continue
		}
		calls++

		st := mapping.MapSynopStation(header, row)
		obs := mapping.MapSynopObservation(header, row)

		if !knownStations[st.Id] {
			storage.InsertStation(db, *st)
			knownStations[st.Id] = true
		}

		batch = append(batch, *obs)

		if len(batch) >= batchSize {
			if err := storage.InsertBatch(db, batch); err != nil {
				slog.Error("insertion batch échouée", "err", err)
				errors++
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		_ = storage.InsertBatch(db, batch)
	}

	slog.Info("ingestion terminée",
		"lignes_traitées", calls,
		"erreurs", errors,
		"durée", time.Since(start).Round(time.Millisecond),
	)
}

func fetch(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 50 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("statut HTTP inattendu: %d", resp.StatusCode)
	}
	return resp, nil
}
