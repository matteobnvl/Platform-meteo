# Ingestion des Données Météorologiques (Stations & Observations)

Ce document décrit en détail le fonctionnement du système d'ingestion des données météorologiques de la plateforme. Ce système récupère les données d'observations synoptiques (SYNOP) fournies en Open Data par Météo-France via la plateforme Data.gouv.fr, les traite, les convertit et les stocke en base de données.

---

## 1. Vue d'ensemble du flux d'ingestion

Le processus d'ingestion est orchestré par le service `ingester` (`cmd/ingester`). Il suit un flux linéaire d'extraction, de transformation et de chargement (ETL) :

```mermaid
flowchart TD
    DG[Data.gouv.fr (Object Storage)] -->|Téléchargement du fichier CSV.GZ| F[Fetcher - fetcher.go]
    F -->|Décompression à la volée| GZ[Gzip Reader]
    GZ -->|Lecture ligne par ligne| CSV[CSV Reader]
    CSV -->|Ligne brute| MAP[Mapping - synop.go]
    
    MAP -->|Mapping Station| MS[model.Station]
    MAP -->|Mapping Observation| MO[model.Observation]
    
    MS -->|Vérification cache local| Cache{Déjà vue dans ce run ?}
    Cache -->|Non| DB_ST[Insert Station - ON CONFLICT DO NOTHING]
    Cache -->|Oui| Skip[Passer l'insertion]
    
    MO -->|Ajout au batch| Batch[Batch de 1000 observations]
    Batch -->|Batch plein ou Fin de fichier| DB_OBS[Insert Batch - Transaction SQL - ON CONFLICT DO NOTHING]
    
    DB_ST --> DB[(Base de données PostgreSQL)]
    DB_OBS --> DB
```

---

## 2. Source de données (Extraction)

Le fetcher récupère les données à partir de l'URL publique de Météo-France hébergée sur Data.gouv.fr.

- **URL ciblée** : `https://object.files.data.gouv.fr/meteofrance/data/synchro_ftp/OBS/SYNOP/synop_{ANNÉE}.csv.gz`
- **Fréquence / Périodicité** : Le script cible par défaut l'année en cours (`time.Now().Year()`).
- **Format** : Fichier compressé au format Gzip (`.gz`) contenant un fichier CSV où le séparateur est le point-virgule (`;`).
- **Optimisation mémoire** : Le fichier n'est pas entièrement téléchargé puis écrit sur le disque. Le flux HTTP (`resp.Body`) est directement passé à un lecteur de décompression Gzip (`gzip.NewReader`), lui-même lu ligne par ligne par un lecteur CSV (`csv.Reader`). Cela garantit une empreinte mémoire constante et très faible (Streaming).

---

## 3. Transformation et Mapping des Données

Chaque ligne lue du CSV est convertie en objets Go typés grâce au package `internal/mapping`.

### A. Mapping des Stations (`MapSynopStation`)
Le fichier contient des données géographiques sur les stations émettrices pour chaque ligne d'observation.
- **Identifiant** : Le champ `geo_id_wmo` sert de clé primaire unique pour la station.
- **Données géographiques** : Les coordonnées `lat` (latitude) et `lon` (longitude) sont converties en `float64`.
- **Nom** : Le champ `name` représente le nom usuel de la station (ex. "Bordeaux Mérignac").

### B. Mapping des Observations (`MapSynopObservation`)
L'observation météorologique contient les données physiques mesurées.
- **Conversion d'unités et normalisation** :
  - **Température** : Le fichier source fournit la température en **Kelvin** (clé `t`). Elle est convertie en **Celsius** dans le modèle :
    $$T_{\text{Celsius}} = T_{\text{Kelvin}} - 273.15$$
  - **Vitesse du vent** : Elle est fournie en **mètres par seconde** (m/s, clé `ff`). Elle est convertie en **km/h** pour l'application :
    $$\text{Vitesse}_{\text{km/h}} = \text{Vitesse}_{\text{m/s}} \times 3.6$$
  - **Direction du vent** : Fournie en degrés (clé `dd`), mappée directement.
  - **Précipitations** : Quantité de pluie tombée sur la dernière heure en mm (clé `rr1`), mappée directement.
  - **Date** : Le champ `validity_time` (format ISO 8601/UTC) est analysé et converti en type `time.Time` Go.

---

## 4. Insertion en Base de Données et Résolution des Conflits (Chargement)

Le stockage en base de données PostgreSQL utilise des stratégies optimisées pour gérer de gros volumes de données tout en évitant les doublons.

### A. Gestion des Stations
Puisque le fichier CSV contient une ligne par observation horaire pour toutes les stations, les informations d'une même station apparaissent de nombreuses fois.
1. **Cache local en mémoire** : Le fetcher utilise une table de hachage `knownStations := make(map[string]bool)`.
2. **Vérification** : Avant d'insérer une station en base de données, l'ingesteur vérifie s'il l'a déjà traitée dans l'exécution courante.
3. **Requête SQL** : Si la station est inconnue du cache local, elle est insérée avec la requête :
   ```sql
   INSERT INTO stations (id, name, latitude, longitude)
   VALUES ($1, $2, $3, $4)
   ON CONFLICT (id) DO NOTHING;
   ```
   Cette clause garantit que si la station existe déjà historiquement en base de données, la ligne n'est pas modifiée et aucune erreur n'est renvoyée. Elle est ensuite ajoutée au cache local pour éviter des appels SQL redondants.

### B. Gestion des Observations (Batching)
L'insertion individuelle d'observations (des dizaines de milliers de lignes par fichier) serait inefficace et saturerait le réseau et la base de données.
1. **Mise en cache par lots (Batching)** : Les observations sont accumulées dans une tranche (slice) en mémoire.
2. **Seuil d'écriture** : Dès que le lot atteint **1000 observations** (défini par `batchSize = 1000`), le lot est envoyé à la base de données.
3. **Transaction SQL** : Le stockage (`internal/storage/observation.go`) utilise une transaction explicite (`db.Begin()`).
4. **Statement Préparé** : Une requête préparée est générée dans la transaction pour exécuter efficacement les insertions en boucle :
   ```sql
   INSERT INTO observations 
       (station_id, observed_at, temperature, wind_speed, wind_direction, precipitation)
   VALUES 
       ($1, $2, $3, $4, $5, $6)
   ON CONFLICT (station_id, observed_at) DO NOTHING;
   ```
5. **Dédoublonnage unique** : La contrainte d'unicité `unique_observation` sur le couple `(station_id, observed_at)` (créée par la migration `0002_constraints.sql`) permet à la clause `ON CONFLICT DO NOTHING` d'ignorer silencieusement les observations déjà enregistrées. Cela permet de relancer l'ingesteur sur le même fichier sans corrompre ni dupliquer les données historiques.
6. **Validation** : La transaction est validée via `tx.Commit()`. En cas d'erreur de traitement d'une des lignes du lot, `tx.Rollback()` annule l'ensemble du lot.

---

## 5. Composants et Fichiers Impliqués

Voici la liste des fichiers clés participant à l'ingestion :

| Fichier | Rôle / Description |
| :--- | :--- |
| [`cmd/ingester/main.go`](file:///Users/matteo/GolandProjects/awesomeProject/platform-meteo/cmd/ingester/main.go) | Point d'entrée de l'application ingesteur. Initialise la base de données (et applique les migrations) avant de lancer le processus d'import. |
| [`internal/fetcher/fetcher.go`](file:///Users/matteo/GolandProjects/awesomeProject/platform-meteo/internal/fetcher/fetcher.go) | Gère la requête HTTP, la décompression Gzip, le streaming du CSV, le cache des stations, le découpage en batches et l'écriture. |
| [`internal/mapping/synop.go`](file:///Users/matteo/GolandProjects/awesomeProject/platform-meteo/internal/mapping/synop.go) | Contient les structures intermédiaires de décodage CSV et les fonctions de conversion d'unités (Kelvin $\rightarrow$ Celsius, m/s $\rightarrow$ km/h). |
| [`internal/storage/station.go`](file:///Users/matteo/GolandProjects/awesomeProject/platform-meteo/internal/storage/station.go) | Exécute l'insertion unitaire de la station avec protection contre les conflits de clé primaire. |
| [`internal/storage/observation.go`](file:///Users/matteo/GolandProjects/awesomeProject/platform-meteo/internal/storage/observation.go) | Gère l'insertion optimisée par transaction et par lots des observations, avec évitement des doublons temporels. |
| [`db/migrations/0002_constraints.sql`](file:///Users/matteo/GolandProjects/awesomeProject/platform-meteo/db/migrations/0002_constraints.sql) | Ajoute la contrainte d'unicité `UNIQUE(station_id, observed_at)` indispensable au bon fonctionnement du dédoublonnage de l'ingesteur. |
