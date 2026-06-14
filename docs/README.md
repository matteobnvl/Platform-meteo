# Documentation de la Plateforme Météo

Bienvenue dans le dossier de documentation du projet. Vous trouverez ici des guides détaillés sur l'architecture et le fonctionnement des différents modules de l'application.

## Sommaire de la Documentation

1. **[Ingestion des Données Météorologiques (Stations et Observations)](./ingestion.md)**
   Ce guide explique en détail comment le service d'ingestion (`cmd/ingester`) récupère, décompresse, convertit et stocke les données d'observations synoptiques (SYNOP) provenant de Météo-France (Data.gouv.fr) en base de données PostgreSQL.
2. **[Flux de Données Météo-France (Data.gouv.fr)](./data_gouv.md)**
   Ce guide récapitule les différents flux de données Météo-France (SYNOP, horaires, bouées, etc.), la correspondance complète des clés et des unités (Kelvin, Celsius, m/s, km/h) ainsi que le schéma conceptuel de la base de données.
3. **[Système de Migrations SQL "Maison"](./migrations.md)**
   Ce guide présente le fonctionnement du moteur de migration Go personnalisé, comment sont ordonnées et jouées les migrations, ainsi que la procédure pour créer une nouvelle migration.
4. **[Schéma et Modèle de la Base de Données](./database.md)**
   Ce guide présente en détail le schéma PostgreSQL (diagramme de base de données, tables, clés, contraintes d'unicité et indexation) ainsi que sa correspondance exacte avec les modèles Go (`structs`).
5. **[Détecteur d'Événements Météorologiques](./detector.md)**
   Ce guide explique le fonctionnement du moteur de détection automatique d'événements (tempêtes, canicules, vagues de froid, inondations) : configuration des seuils, algorithmes de détection, déduplication et endpoints API associés.




## Comment lancer l'Ingestion ?

Le service d'ingestion est configuré dans le fichier `compose.yml` et s'exécute avec l'outil de live-reload `air` pour le développement.

### Prérequis
1. S'assurer d'avoir un fichier `.env` configuré à la racine du projet (voir le guide d'installation dans le [README.md principal](../README.md)).
2. Lancer les services avec Docker Compose :
   ```bash
   make up
   ```

### Lancement effectif de l'ingestion
> [!NOTE]
> Dans le code actuel de [`cmd/ingester/main.go`](../cmd/ingester/main.go), l'appel à la fonction d'ingestion `fetcher.FetchSynop(db)` est commenté.
> Pour déclencher l'ingestion lors du démarrage du conteneur `ingester` :
> 1. Décommentez la ligne `// fetcher.FetchSynop(db)` dans [`cmd/ingester/main.go`](../cmd/ingester/main.go).
> 2. Le conteneur se rechargera automatiquement grâce à `air`.
