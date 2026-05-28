# Platforme-meteo


## Chargement des données

Le fichier `stations_data.json` contient les 30 stations qui seront surveillées sur la plateforme météo. L'idée est de 
charger les 30 stations en base au démarrage de l'application, si elles n'existent pas en base.
Ensuite, avec le ingester nous appelerons cette route API (qui fonctionne uniquement avec les données de lattitude et 
de longitude pour récupérer des données météo) :

```
https://api.open-meteo.com/v1/forecast?latitude=44.8333&longitude=-0.7&start_date=2026-04-01&end_date=2026-05-01&hourly=temperature_2m,wind_gusts_10m,precipitation&timezone=auto
```
Petite subtilité de l'API, par défaut elle retourne les données sur les 7 derniers jours. Si on veut une plage plus 
grande il est conseillé de passer par `archive-api.open-meteo.com` qui permet de récupérer des données plus anciennes 
et plus rapidement.

Ici, on vient appeler l'API de l'open-meteo avec les coordonnées de Bordeaux Mérignac. 

Tableau des query utilisées :

| Paramètre        | Valeur dans ton URL                                       | Utilité et Description                                                                                                                                                                                                                                                             |
|:-----------------|:----------------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`latitude`**   | `44.8333`                                                 | **Coordonnée Nord/Sud** du point géographique. Ici, cela correspond à la latitude de Bordeaux Mérignac.                                                                                                                                                                            |
| **`longitude`**  | `-0.7`                                                    | **Coordonnée Est/Ouest** du point géographique. Le signe moins `-` indique qu'on est à l'Ouest du méridien de Greenwich (Bordeaux).                                                                                                                                                |
| **`start_date`** | `2026-04-01`                                              | **Date de début** de la plage de données souhaitée (au format `AAAA-MM-JJ`).                                                                                                                                                                                                       |
| **`end_date`**   | `2026-05-01`                                              | **Date de fin** de la plage de données souhaitée (au format `AAAA-MM-JJ`). Associé à `start_date`, cela définit une fenêtre de 30 jours.                                                                                                                                           |
| **`hourly`**     | `temperature_2m,`<br>`wind_gusts_10m,`<br>`precipitation` | **Variables météo demandées heure par heure** :<br>• `temperature_2m` : Température à 2m (canicules/vagues de froid).<br>• `wind_gusts_10m` : Rafales de vent à 10m (tempêtes).<br>• `precipitation` : Cumul de pluie en mm (inondations/sécheresses).                             |
| **`timezone`**   | `auto`                                                    | **Gestion du fuseau horaire**. L'option `auto` demande à Open-Meteo de caler automatiquement les heures du JSON sur le fuseau de la coordonnée demandée (ex: `Europe/Paris` pour Bordeaux), évitant ainsi les décalages liés aux heures d'été/hiver lors du parsing en Go.         |


Nous récupérons un format comme ceci :

```
(Racine du JSON)
├── latitude (float64) -> 44.83
├── longitude (float64) -> -0.6999998
├── generationtime_ms (float64)
├── utc_offset_seconds (int)
├── timezone (string) -> "GMT"
├── timezone_abbreviation (string) -> "GMT"
├── elevation (int) -> 49
│
├── hourly_units (Objet)
│   ├── time (string) -> "iso8601"
│   └── temperature_2m (string) -> "°C"
│
└── hourly (Objet - Format Colonne)
├── time (Slice string) -> ["2026-05-28T00:00", "2026-05-28T01:00", ...]
└── temperature_2m (Slice float64) -> [22, 20.8, 19.7, ...]
```

| Chemin dans le JSON           | Type             | Exemple de valeur           | Description                                                   |
|:------------------------------|:-----------------|:----------------------------|:--------------------------------------------------------------|
| `latitude`                    | `Number (float)` | `44.83`                     | Latitude géographique du résultat renvoyé.                    |
| `longitude`                   | `Number (float)` | `-0.6999998`                | Longitude géographique du résultat renvoyé.                   |
| `generationtime_ms`           | `Number (float)` | `0.532`                     | Temps mis par le serveur Open-Meteo pour générer la réponse.  |
| `utc_offset_seconds`          | `Number (int)`   | `0`                         | Décalage en secondes par rapport au temps UTC.                |
| `timezone`                    | `String`         | `"GMT"`                     | Fuseau horaire appliqué.                                      |
| `timezone_abbreviation`       | `String`         | `"GMT"`                     | Abréviation textuelle du fuseau.                              |
| `elevation`                   | `Number (int)`   | `49`                        | Altitude en mètres par rapport au niveau de la mer.           |
| `hourly_units`                | `Object`         | `{ ... }`                   | Conteneur des unités de mesure.                               |
| `hourly_units.time`           | `String`         | `"iso8601"`                 | Format utilisé pour les dates/heures.                         |
| `hourly_units.temperature_2m` | `String`         | `"°C"`                      | Unité utilisée pour la température.                           |
| `hourly`                      | `Object`         | `{ ... }`                   | **Le bloc de données temporelles (Slices synchronisées).**    |
| `hourly.time`                 | `Array [String]` | `["2026-05-28T00:00", ...]` | Liste ordonnée de tous les timestamps (heure par heure).      |
| `hourly.temperature_2m`       | `Array [Float]`  | `[22, 20.8, 19.7, ...]`     | Liste ordonnée des températures correspondant à chaque heure. |