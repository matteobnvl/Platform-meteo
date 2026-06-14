# Platform Météo

Une plateforme de traitement et d'exposition des données météorologiques de Météo-France (Data.gouv.fr).

Elle ingère les observations SYNOP en continu, les stocke en PostgreSQL, détecte les événements météo sensibles (tempêtes, canicules, vagues de froid, inondations) et les expose via une API REST.

---

## 📖 Documentation interne

Une documentation détaillée est disponible dans le dossier [**`/docs`**](./docs/README.md) :
- 📥 **[Ingestion des Données](./docs/ingestion.md)** : Flux ETL complet, batching, gestion des doublons.
- 🗄️ **[Schéma de la Base de Données](./docs/database.md)** : Tables SQL, clés, index et correspondance avec les modèles Go.
- ⚙️ **[Système de Migrations](./docs/migrations.md)** : Runner de migration Go maison.
- 🌐 **[Flux Data.gouv.fr](./docs/data_gouv.md)** : Source des données, clés CSV et conversions d'unités.

---

## Architecture

Le projet est découpé en deux binaires indépendants :

```
cmd/
├── api/        → serveur HTTP (lecture seule sur la BDD)
└── ingester/   → ingestion SYNOP + détection d'événements

internal/
├── handler/    → handlers HTTP (stations, observations, events)
├── fetcher/    → téléchargement et parsing du CSV SYNOP
├── mapping/    → conversion des données brutes vers le modèle interne
├── detector/   → détection des événements météo sensibles
├── model/      → types Go neutres (Station, Observation, Event)
└── storage/    → insertions BDD (batch observations, stations)

db/
├── connexion.go        → connexion PostgreSQL
├── query.go            → requêtes stations et observations
├── events.go           → requêtes événements
└── migrations/         → fichiers SQL versionnés + runner
```

**Flux de l'ingester :**
1. Télécharge le fichier SYNOP annuel depuis data.gouv.fr (CSV.GZ streamé, pas chargé en mémoire)
2. Mappe chaque ligne vers les types internes (conversions K→°C, m/s→km/h)
3. Insère les stations + observations en batch de 1000 dans PostgreSQL
4. Lance la détection d'événements sur toutes les stations en parallèle (une goroutine par station)

L'ingestion se répète **toutes les 24h**, la détection **toutes les 10 minutes**.

---

## Installation

Copier le fichier d'exemple et remplir les variables :

```bash
cp .env.example .env
```

Puis lancer tout avec :

```bash
make up
```

Les trois conteneurs (postgres, api, ingester) démarrent automatiquement via Docker Compose.

---

## Variables d'environnement

| Variable | Description | Défaut |
|---|---|---|
| `DB_HOST` | Hôte PostgreSQL | `postgres` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | Utilisateur BDD | — |
| `DB_PASSWORD` | Mot de passe BDD | — |
| `DB_NAME` | Nom de la base | `meteo_db` |
| `STORM_WIND_THRESHOLD` | Seuil vent tempête (km/h) | `80` |
| `STORM_DURATION_HOURS` | Durée min tempête (h) | `3` |
| `HEATWAVE_TEMP_THRESHOLD` | Seuil canicule (°C) | `35` |
| `HEATWAVE_DURATION_DAYS` | Durée min canicule (jours) | `3` |
| `COLDWAVE_TEMP_THRESHOLD` | Seuil vague de froid (°C) | `0` |
| `COLDWAVE_DURATION_DAYS` | Durée min vague de froid (jours) | `3` |
| `FLOOD_PRECIP_THRESHOLD` | Seuil inondation (mm/24h) | `50` |

---

## Commandes utiles

```bash
make up               # démarre tout
make down             # arrête tout
make rebuild          # reconstruit et redémarre
make logs-api         # logs du serveur API
make logs-ingester    # logs de l'ingester
```

---

## Migrations

Les migrations sont versionnées dans `db/migrations/` et s'appliquent automatiquement au démarrage de l'ingester.

Pour repartir de zéro :

```bash
make down
docker volume rm platform-meteo_postgres_data_dev
make up
```

---

## API REST

Le serveur écoute sur le port **8080**.

### Stations

| Méthode | Route | Description |
|---|---|---|
| `GET` | `/stations` | Liste toutes les stations |
| `GET` | `/stations?country=FR` | Filtre par code pays |
| `GET` | `/stations/{id}` | Détail d'une station |
| `POST` | `/stations` | Crée une station |
| `PUT` | `/stations/{id}` | Met à jour une station |
| `DELETE` | `/stations/{id}` | Supprime une station |

**Exemple — liste des stations françaises :**
```bash
curl http://localhost:8080/stations?country=FR
```

**Exemple — créer une station :**
```bash
curl -X POST http://localhost:8080/stations \
  -H "Content-Type: application/json" \
  -d '{"Id":"07149","Name":"Paris-Montsouris","Country":"FR","Latitude":48.8,"Longitude":2.3}'
```

---

### Observations

| Méthode | Route | Description |
|---|---|---|
| `GET` | `/stations/{id}/observations` | Observations d'une station |
| `GET` | `/stations/{id}/observations?from=2026-01-01&to=2026-01-31` | Filtre par intervalle de dates |
| `GET` | `/stations/{id}/observations?limit=100&offset=0` | Pagination |
| `GET` | `/observations/aggregate?period=daily&station_id={id}` | Agrégation journalière (avg/max/min temp) |

**Exemple :**
```bash
curl "http://localhost:8080/stations/07149/observations?from=2026-01-01&to=2026-01-31&limit=50"
```

---

### Événements météo

| Méthode | Route | Description |
|---|---|---|
| `GET` | `/events` | Liste tous les événements |
| `GET` | `/events?type=storm` | Filtre par type |
| `GET` | `/events?from=2026-01-01&to=2026-01-31` | Filtre par période |
| `GET` | `/events/{id}` | Détail d'un événement |
| `GET` | `/events/stats?type=storm&country=FR` | Comptage agrégé par type/pays |
| `GET` | `/stations/{id}/events` | Événements d'une station |

Types d'événements détectés : `storm`, `heatwave`, `cold_wave`, `flood`

**Exemple :**
```bash
curl "http://localhost:8080/events?type=storm&from=2026-01-01"
curl "http://localhost:8080/events/stats?country=FR"
```

---

### Santé

```bash
curl http://localhost:8080/health
```

---

## Format des erreurs

Toutes les erreurs sont renvoyées en JSON avec le code HTTP approprié :

```json
{
  "error": "station \"07999\" introuvable"
}
```

---

## Système de migrations

Le projet intègre un runner de migration SQL maison dans `db/migrations/migration.go`.

Les fichiers SQL sont appliqués dans l'ordre numérique au démarrage de l'ingester, et chaque migration est tracée dans une table `migrations` pour ne pas être réappliquée.

Pour créer une nouvelle migration, il suffit d'ajouter un fichier `000X_nom.sql` dans `db/migrations/`.
