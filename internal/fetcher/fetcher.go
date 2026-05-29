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
)

func FetchSynop(db *sql.DB) {
	url := fmt.Sprintf(
		"https://object.files.data.gouv.fr/meteofrance/data/synchro_ftp/OBS/SYNOP/synop_%d.csv.gz",
		time.Now().Year(),
	)

	resp := fetch(url)
	defer resp.Body.Close()

	header, csvReader := readCsv(resp)
	var total int
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		total++

		obs := mapping.MapSynopObservation(header, row)
		st := mapping.MapSynopStation(header, row)
		fmt.Println(obs, st)
	}

	fmt.Printf("ingestion terminée de %d observations \n", total)
}

func fetch(url string) *http.Response {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("statut %d", resp.StatusCode)
	}
	return resp
}

func readCsv(resp *http.Response) ([]string, *csv.Reader) {
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("erreur décompression", err)
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		log.Fatal("erreur lecture header", err)
	}

	return header, csvReader
}
