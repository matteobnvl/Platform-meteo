package fetcher

import (
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
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

	resp := fetch(url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Statut HTTP invalide: %d", resp.StatusCode)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("Erreur décompression gzip: %v", err)
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	// lecture du header
	header, err := csvReader.Read()
	if err != nil {
		log.Fatalf("Erreur lecture header: %v", err)
	}

	// stock les bases connues pour pas dupliquer l'insertion des bases
	knownStations := make(map[string]bool)

	var batch []model.Observation

	log.Println("début de l'importation des données...")

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("erreur de lecture d'une ligne: %v", err)
			continue
		}

		// mapping
		st := mapping.MapSynopStation(header, row)
		obs := mapping.MapSynopObservation(header, row)

		// si pas de stations, on insert
		if !knownStations[st.Id] {
			storage.InsertStation(db, *st)
			knownStations[st.Id] = true
		}

		batch = append(batch, *obs)

		if len(batch) >= batchSize {
			err := storage.InsertBatch(db, batch)
			if err != nil {
				log.Printf("Erreur lors de l'insertion du batch: %v", err)
			}
			batch = batch[:0]
		}
	}

	// insertions des dernières observations
	if len(batch) > 0 {
		_ = storage.InsertBatch(db, batch)
	}

	log.Println("Importation terminée avec succès !")
}

func fetch(url string) *http.Response {
	client := &http.Client{Timeout: 50 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("statut %d", resp.StatusCode)
	}
	return resp
}
