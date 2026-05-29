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
)

func Run(db *sql.DB) {
	url := fmt.Sprintf(
		"https://object.files.data.gouv.fr/meteofrance/data/synchro_ftp/OBS/SYNOP/synop_%d.csv.gz",
		time.Now().Year(),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("statut %d", resp.StatusCode)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("erreur décompression", err)
		return
	}
	defer gzReader.Close()

	csvReader := csv.NewReader(gzReader)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	_, err = csvReader.Read()
	if err != nil {
		log.Fatal("erreur lecture header", err)
		return
	}

	var total int
	for {
		_, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		total++
	}

	fmt.Printf("ingestion terminée de %d observations \n", total)
}
