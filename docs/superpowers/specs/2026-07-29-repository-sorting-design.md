# Repository Sorting Design

## Goal

Add deterministic sorting to `GET /v1/repositories`, based on the approach used by `developer-overheid-nl/don-api-register#184`, and include every currently open Dependabot version update that still changes this repository.

## API contract

`GET /v1/repositories` accepts two optional query parameters:

- `sortBy=title|lastActivity`
- `sortOrder=asc|desc`

Omitting either parameter applies its default independently: `sortBy=title` and `sortOrder=asc`. `title` maps to `Repository.Name`; `lastActivity` maps only to `Repository.LastActivityAt` and never to `Repository.LastCrawledAt`.

Unsupported values return HTTP 400 with `application/problem+json`. The invalid parameter is identified as a query parameter named `sortBy` or `sortOrder`.

Sorting applies to the combined result of the existing filters and optional `q` search term. It happens before pagination. The deprecated `/v1/repositories/_search` endpoint is unchanged.

Pagination links preserve `sortBy` and `sortOrder` because they are built from the request query string.

## Sorting behavior

Title comparison is case-insensitive. Last-activity comparison is chronological. A zero `LastActivityAt` represents a missing value and is always placed after repositories with an activity timestamp, for both ascending and descending order.

Results are deterministic across pages. Equal primary values use case-insensitive title and then repository ID as ascending tie-breakers. For equal titles, repository ID is the final ascending tie-breaker.

## Architecture and data flow

A typed repository-sort model owns the accepted fields, directions, defaults, and validation errors. The service parses the HTTP query values before calling the repository. Invalid input therefore stops before data access.

The repository receives the parsed sort value. The existing list path loads active repositories and applies filters in memory; the filtered slice is therefore sorted in memory immediately before pagination. A focused sorting helper keeps comparison rules separate from filtering and persistence.

The handler continues to use the shared pagination-header helper. No additional pagination-link implementation is needed.

## OpenAPI and changelog

The checked-in `api/openapi.json` documents both parameters as reusable query parameters, including enums and defaults, and references them from `GET /repositories`. The endpoint description states that filtered results are sorted before pagination.

An unreleased Changie fragment records title and last-activity sorting in both directions.

The resulting OAS must:

- parse as JSON;
- pass the repository's OpenAPI contract tests;
- pass `@developer-overheid-nl/don-checker` with ruleset `adr-21` and `--fail-on error`.

## Dependabot updates

The open Dependabot PR inventory at design time is:

- #106: `github.com/go-playground/validator/v10` 10.30.2 to 10.30.3. The feature branch already contains 10.30.3 through `main`, with newer compatible indirect `golang.org/x/*` versions, so no older lockfile patch is applied.
- #138: pin `actions/checkout` 7.0.1 in `.github/workflows/oas-adr.yml`.
- #139: pin `actions/setup-node` 7.0.0 in `.github/workflows/oas-adr.yml`.
- #140: pin `docker/login-action` 4.5.2 in the production deploy, test deploy, and Go CI workflows.

All action updates remain pinned to the exact commit SHAs supplied by their Dependabot PRs.

## Testing

Tests cover:

- parsing defaults, supported values, and invalid values;
- case-insensitive title sorting in both directions;
- chronological last-activity sorting in both directions;
- missing activity timestamps placed last;
- deterministic tie-breakers;
- sorting after filtering and before pagination;
- service forwarding and validation without repository access on invalid input;
- HTTP status, problem details, pagination links, and end-to-end ordering;
- OpenAPI parameter references, enums, and defaults.

The full Go suite, formatting checks, workflow syntax inspection, JSON parsing, and DON Checker validation form the completion gate.

## Out of scope

This change does not add sorting to organisations, Git organisations, filter options, or the deprecated repository search endpoint. It does not change crawler timestamps, data ingestion, or Typesense indexing.
