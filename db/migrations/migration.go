package migrations

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
)

func Run(db *sql.DB) error {
	// création de la table migrations
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
		    filename TEXT PRIMARY KEY,
		    applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)

	if err != nil {
		return fmt.Errorf("impossible de créer la table migrations : %w", err)
	}

	// lecture des fichiers dans /migrations
	files, err := filepath.Glob("./db/migrations/*.sql")
	if err != nil {
		return fmt.Errorf("impossible de lire les fichiers de migrations : %w", err)
	}
	sort.Strings(files) // trie par 001, 002

	// application des migrations si pas exécuté
	for _, file := range files {
		filename := filepath.Base(file)
		var count int
		db.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = $1", filename).Scan(&count)
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("impossible de lire %s : %w", file, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("erreur dans %s : %w", file, err)
		}

		db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
		slog.Info("migration appliquée", "fichier", filename)
	}
	return nil
}
