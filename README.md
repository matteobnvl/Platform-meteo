# PLatform M&t&o

## Système de migrations

### Fonctionnement
Le projet contient un système de migrations maison, présent dans le dossier `db/migrations/`.
Un fichier `migration.go` :
- crée la table de migrations si elle n'existe pas
  - récupère tous les fichiers de migration présent dans le dossier
  - vérifie si la table de migrations existe
  - si elle n'existe pas, elle exécute le code dans le fichier et insère en base la migration

### Créer une migration

Créer un fichier dans `db/migrations/` en suivant le nommage :
```text
0001_init.sql
0002_example.sql
```

Attention de bien respecter le suivi des nombres pour ne pas casser les migrations (un sort est effectué sur ces chiffres).

Dans le fichier, écrire le code sql qui doit être exécuté.

Automatiquement le fichier sera lu, les migrations sont exécuté à chaque boot du binaire (dans `cmd/api/main.go`)

### Règles importantes

1. Ne jamais modifier un fichier existant, toujours en créer un nouveau
2. Ne jamais sauter de numéro
3. Un fichier = une modification précise