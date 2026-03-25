# OSS-register API

API voor het OSS-register van developer.overheid.nl.

## Overzicht

- API-versie: `1.0.0`
- Lokale base URL: `http://localhost:1337/v1`
- OpenAPI-documentatie: `http://localhost:1337/v1/openapi.json`

## Vereisten

- Go `1.25`
- Docker met `docker compose`

## Lokaal draaien

1. Start PostgreSQL:

   ```bash
   docker compose up -d
   ```

2. Controleer de omgevingsvariabelen in `.env`:

   ```dotenv
   DB_HOSTNAME=localhost
   DB_USERNAME=don
   DB_PASSWORD=don
   DB_DBNAME=don_oss_v1
   DB_SCHEMA=public
   CRAWL_STALE_AFTER_HOURS=48
   ```

3. Start de API:

   ```bash
   go run ./cmd
   ```

De server luistert standaard op poort `1337`. Bij het opstarten worden database-migraties automatisch uitgevoerd.

## Handige commands

```bash
go test ./...
golangci-lint run --timeout 5m
```

## Changelog (Changie)

Voor user-facing wijzigingen, zoals een fix, feature of breaking change, verwachten we per PR een Changie-fragment in `.changes/unreleased`.

Eenmalig installeren:

```bash
go install github.com/miniscruff/changie@latest
```

Fragment aanmaken:

```bash
changie new
```

Normaal is een fragment niet nodig voor interne refactors zonder zichtbaar effect, docs-only wijzigingen en CI-only tweaks.

Bij een release kun je de fragmenten bundelen in `CHANGELOG.md` met:

```bash
changie batch <version>
```

Dit gebeurt ook automatisch bij elke merge naar `main` via GitHub Actions: eerst `changie batch auto` en daarna `changie merge`. Daarna wordt automatisch een PR aangemaakt met de changelog-updates.

## Database en pgAdmin

De applicatie gebruikt PostgreSQL. De `docker compose`-config start lokaal een Postgres 17-container met deze waarden:

- Host: `localhost`
- Port: `5432`
- Username: `don`
- Password: `don`
- Database: `don_oss_v1`

Voor het beheren van de database kun je optioneel [pgAdmin](https://www.pgadmin.org/) gebruiken:

```bash
docker run --rm -p 5050:80 \
  -e PGADMIN_DEFAULT_EMAIL=admin@example.com \
  -e PGADMIN_DEFAULT_PASSWORD=admin \
  dpage/pgadmin4
```