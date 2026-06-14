# Platform Météo

Une plateforme de traitement et d'exposition des données météorologiques de Météo-France (Data.gouv.fr).

## 📖 Documentation interne

Une documentation détaillée de l'architecture et du fonctionnement interne du projet est disponible dans le dossier [**`/docs`**](./docs/README.md) :
- 📥 **[Ingestion des Données](./docs/ingestion.md)** : Flux d'extraction, de transformation et d'écriture par lots (batching) des stations et des observations.
- 🗄️ **[Schéma de la Base de Données](./docs/database.md)** : Tables SQL, clés, index/contraintes et correspondance avec les modèles structurés Go.
- ⚙️ **[Système de Migrations](./docs/migrations.md)** : Fonctionnement détaillé du runner de migration Go maison.
- 🌐 **[Flux Data.gouv.fr](./docs/data_gouv.md)** : Liste des flux Météo-France, clés API et correspondances d'unités (Kelvin, Celsius, m/s, km/h).

---

## Installation

Faire `cp .env.example .env` puis remplir les variables d'environnement.

Et lancer dans un terminal :
```bash
make up
```

Les applications vont se lancer automatiquement via Docker Compose.

---

## Système de migrations

Le projet intègre un système de migrations SQL personnalisé situé dans le dossier [`db/migrations/`](./db/migrations/).

Pour comprendre son fonctionnement interne ou savoir comment créer une nouvelle migration étape par étape, veuillez vous référer à la **[documentation du système de migrations](./docs/migrations.md)**.