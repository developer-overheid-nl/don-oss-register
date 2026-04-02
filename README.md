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