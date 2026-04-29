# API registratie

API van het OSS register (oss.developer.overheid.nl)

## Overview

- API version: 1.0.0

## Lokaal draaien

1. Start de afhankelijkheden:

   ```bash
   docker compose up -d
   ```

2. Start de server:

   ```bash
   go run cmd/main.go
   ```

   De API luistert standaard op poort **1337**.

## Typesense integratie

Nieuwe repositories worden na een succesvolle POST of PUT ook naar Typesense gestuurd, zodat ze vindbaar zijn in de zoekfunctie. Bij het opstarten van de server worden bestaande actieve repositories bovendien opnieuw naar Typesense gepubliceerd. Stel hiervoor de volgende omgevingsvariabelen in:

- `TYPESENSE_ENDPOINT`: basis-URL van de Typesense cluster (bijv. `https://search.don.apps.digilab.network`).
- `TYPESENSE_API_KEY`: API key met schrijfrechten.
- `TYPESENSE_COLLECTION`: naam van de collectie (standaard `oss-register`).
- `TYPESENSE_DETAIL_BASE_URL`: basis-URL voor detailpagina's in de frontend (standaard `https://oss.developer.overheid.nl/repositories`).
- `ENABLE_TYPESENSE`: zet op `false` om Typesense indexing volledig uit te schakelen (standaard `true`).

## Database en pgAdmin

De applicatie gebruikt PostgreSQL. De docker-compose start automatisch een Postgres container met bovenstaande credentials.

Voor het beheren van de database kun je optioneel [pgAdmin](https://www.pgadmin.org/) gebruiken:

```bash
docker run --rm -p 5050:80 \
  -e PGADMIN_DEFAULT_EMAIL=admin@example.com \
  -e PGADMIN_DEFAULT_PASSWORD=admin \
  dpage/pgadmin4
```

Navigeer naar `http://localhost:5050`, voeg een nieuwe server toe en gebruik de waarden:

- Host: `localhost`
- Port: `5432`
- Username: `don`
- Password: `don`
- Database: `don_v1`

## Data synchroniseren tussen omgevingen

Met `scripts/sync_organisations.sh` kun je organisaties en git-organisaties van
een bronomgeving naar een doelomgeving pompen. Dit is handig om bijvoorbeeld
data uit productie of acceptatie over te zetten naar een testomgeving.

Het script leest uit de bron:

- `GET /organisations`
- `GET /git-organisations`

en schrijft naar het doel:

- `POST /organisations`
- `POST /git-organisations`

Gebruik basis-URL's inclusief `/oss-register/v1`.

```bash
SOURCE_BASE_URL="https://api.developer.overheid.nl/oss-register/v1" \
TARGET_BASE_URL="https://api.don.projects.digilab.network/oss-register/v1" \
SOURCE_API_KEY="..." \
TARGET_BEARER_TOKEN="..." \
./scripts/sync_organisations.sh
```

Voor de bron kun je `SOURCE_API_KEY` of `SOURCE_BEARER_TOKEN` gebruiken. Voor
het doel is `TARGET_BEARER_TOKEN` verplicht. Het script volgt paginering via de
`Link` header en schrijft records die niet verwerkt konden worden naar
`sync-organisations-errors.json`.

Handige opties:

- `PER_PAGE=100`: aantal records per pagina.
- `SLEEP_SECONDS=0`: pauze tussen POST requests.
- `OUT=sync-organisations-errors.json`: pad voor het foutenbestand.
- `SKIP_ORGANISATIONS=1`: alleen git-organisaties synchroniseren.
- `SKIP_GIT_ORGANISATIONS=1`: alleen organisaties synchroniseren.

Git-organisaties verwijzen naar een bestaande organisatie via
`organisation.uri`. Synchroniseer organisaties dus eerst, tenzij die al in de
doelomgeving bestaan.

## Deployen

De deployment van deze site verloopt via GitHub Actions en een aparte infra
repository.

### Benodigde variabelen en secrets

- Organization variable `INFRA_REPO`, bijvoorbeeld
  `developer-overheid-nl/don-infra`.
- Repository variable `KUSTOMIZE_PATH`, met als basispad bijvoorbeeld
  `apps/api/overlays/`.
- Secrets `RELEASE_PROCES_APP_ID` en `RELEASE_PROCES_APP_PRIVATE_KEY` voor het
  aanpassen van de infra repository.

### Deploy naar test

De testdeploy draait via
`.github/workflows/deploy-test.yml`.

- De workflow draait op pushes naar branches behalve `main`.
- Alleen commits met `[deploy-test]` in de commit message worden echt gedeployed.
- Er wordt een image gebouwd en gepusht naar
  `ghcr.io/<owner>/<repo>` met tags `test` en de commit SHA.
- Daarna wordt in `INFRA_REPO` het bestand
  `${KUSTOMIZE_PATH}test/kustomization.yaml` bijgewerkt naar de nieuwe image
  tag en direct gecommit.

Voorbeeld commit message:

```text
feat: pas content aan [deploy-test]
```

### Deploy naar productie

De productiedeploy draait via
`.github/workflows/deploy-prod.yml`.

- De workflow draait bij een push naar `main`.
- Er wordt in `INFRA_REPO` een release branch aangemaakt.
- In `${KUSTOMIZE_PATH}prod/kustomization.yaml` wordt de image tag bijgewerkt
  naar de commit SHA van deze repository.
- Daarna wordt automatisch een pull request in de infra repository geopend.
- De productie-uitrol gebeurt door die pull request te mergen.

### Contributies en deploy

Een contribution of pull request leidt niet automatisch tot een deployment.

- Een pull request triggert wel CI, waaronder de build en JSON-validatie.
- De build in `.github/workflows/go-ci.yml` bouwt voor een pull request een
  Docker image als controle, maar pusht dat image niet naar GHCR en past de
  infra repository niet aan.
- Er is dus geen automatische preview-omgeving per pull request.
- Een testdeploy gebeurt pas na een push naar een branch in deze repository met
  `[deploy-test]` in de commit message.
- Die testdeploy gebruikt repository- en organization-variables en secrets om
  ook `INFRA_REPO` aan te passen. Daardoor is dit pad in de praktijk bedoeld
  voor maintainers of contributors met een branch in deze repository.
