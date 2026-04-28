#!/usr/bin/env bash
set -euo pipefail

# Sync organisations en git-organisations van een bronomgeving naar een doelomgeving.

SOURCE_BASE_URL="${SOURCE_BASE_URL:-https://api.developer.overheid.nl/oss-register/v1}"
TARGET_BASE_URL="${TARGET_BASE_URL:-https://api.don.projects.digilab.network/oss-register/v1}"

SOURCE_API_KEY=""
TARGET_BEARER_TOKEN=""

PER_PAGE="${PER_PAGE:-100}"
SLEEP_SECONDS="${SLEEP_SECONDS:-0}"
OUT="${OUT:-sync-organisations-errors.json}"
SKIP_ORGANISATIONS="${SKIP_ORGANISATIONS:-0}"
SKIP_GIT_ORGANISATIONS="${SKIP_GIT_ORGANISATIONS:-0}"

org_success=0
org_failed=0
org_skipped=0
git_org_success=0
git_org_failed=0
git_org_skipped=0

declare -a SOURCE_AUTH_ARGS=()
declare -a TARGET_AUTH_ARGS=()

usage() {
  cat <<'EOF'
Usage:
  SOURCE_BASE_URL="https://source.example.nl/v1" \
  TARGET_BASE_URL="https://target.example.nl/v1" \
  SOURCE_API_KEY="..." \
  TARGET_BEARER_TOKEN="..." \
  ./scripts/sync_organisations.sh

Of met positional args:
  ./scripts/sync_organisations.sh <source_base_url> <target_base_url>

Belangrijk:
  - Gebruik basis-URL's inclusief /v1.
  - Voor de bron kun je SOURCE_API_KEY of SOURCE_BEARER_TOKEN gebruiken.
  - Voor het doel is TARGET_BEARER_TOKEN of BEARER_TOKEN verplicht.
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Vereiste command ontbreekt: $1" >&2
    exit 1
  fi
}

normalise_bearer_token() {
  local token="${1:-}"

  token="${token#Bearer }"
  token="${token#bearer }"
  printf '%s' "$token"
}

build_source_auth_args() {
  SOURCE_AUTH_ARGS=()
  SOURCE_BEARER_TOKEN="$(normalise_bearer_token "$SOURCE_BEARER_TOKEN")"

  if [[ -n "$SOURCE_BEARER_TOKEN" ]]; then
    SOURCE_AUTH_ARGS+=(-H "Authorization: Bearer ${SOURCE_BEARER_TOKEN}")
  elif [[ -n "$SOURCE_API_KEY" ]]; then
    SOURCE_AUTH_ARGS+=(-H "X-Api-Key: ${SOURCE_API_KEY}")
  fi
}

build_target_auth_args() {
  TARGET_AUTH_ARGS=()
  TARGET_BEARER_TOKEN="$(normalise_bearer_token "$TARGET_BEARER_TOKEN")"

  if [[ -z "$TARGET_BEARER_TOKEN" ]]; then
    echo "TARGET_BEARER_TOKEN ontbreekt; POSTs naar het doel vereisen een bearer token." >&2
    exit 1
  fi

  TARGET_AUTH_ARGS+=(-H "Authorization: Bearer ${TARGET_BEARER_TOKEN}")
}

prompt_source_auth() {
  if [[ ! -t 0 ]]; then
    echo "Bron-auth ontbreekt of is ongeldig. Zet SOURCE_API_KEY of SOURCE_BEARER_TOKEN." >&2
    exit 1
  fi

  echo
  read -r -p "Bron X-Api-Key (leeg voor bearer token): " SOURCE_API_KEY
  if [[ -z "$SOURCE_API_KEY" ]]; then
    read -r -s -p "Bron Bearer token: " SOURCE_BEARER_TOKEN
    SOURCE_BEARER_TOKEN="$(normalise_bearer_token "$SOURCE_BEARER_TOKEN")"
    echo
  fi
}

prompt_target_auth() {
  if [[ ! -t 0 ]]; then
    echo "Doel-auth ontbreekt of is ongeldig. Zet TARGET_BEARER_TOKEN of BEARER_TOKEN." >&2
    exit 1
  fi

  echo
  read -r -s -p "Doel Bearer token: " TARGET_BEARER_TOKEN
  TARGET_BEARER_TOKEN="$(normalise_bearer_token "$TARGET_BEARER_TOKEN")"
  echo
}

request_source_once() {
  local url="$1"
  local body_file="$2"
  local headers_file="$3"
  local http_code
  local -a curl_args

  build_source_auth_args
  curl_args=(
    -sS
    -D "$headers_file"
    -o "$body_file"
    -w '%{http_code}'
    "$url"
    -H 'Accept: application/json'
  )
  if ((${#SOURCE_AUTH_ARGS[@]} > 0)); then
    curl_args+=("${SOURCE_AUTH_ARGS[@]}")
  fi
  http_code=$(curl "${curl_args[@]}")

  printf '%s' "$http_code"
}

request_target_once() {
  local url="$1"
  local payload="$2"
  local body_file="$3"
  local http_code
  local -a curl_args

  build_target_auth_args
  curl_args=(
    -sS
    -o "$body_file"
    -w '%{http_code}'
    -X POST "$url"
    -H 'Accept: application/json'
    -H 'Content-Type: application/json'
    -d "$payload"
  )
  curl_args+=("${TARGET_AUTH_ARGS[@]}")
  http_code=$(curl "${curl_args[@]}")

  printf '%s' "$http_code"
}

append_error() {
  local phase="$1"
  local descriptor="$2"
  local request_json="$3"
  local status="$4"
  local body_file="$5"

  if jq -e . >/dev/null 2>&1 < "$body_file"; then
    jq \
      --arg phase "$phase" \
      --arg descriptor "$descriptor" \
      --argjson request "$request_json" \
      --argjson status "$status" \
      --slurpfile body "$body_file" \
      '. + [{phase:$phase, descriptor:$descriptor, request:$request, status:$status, body:$body[0]}]' \
      "$OUT" > "${OUT}.tmp"
  else
    jq \
      --arg phase "$phase" \
      --arg descriptor "$descriptor" \
      --argjson request "$request_json" \
      --arg status "$status" \
      --rawfile body "$body_file" \
      '. + [{
        phase:$phase,
        descriptor:$descriptor,
        request:$request,
        status:(if ($status | test("^[0-9]+$")) then ($status|tonumber) else $status end),
        body:$body
      }]' \
      "$OUT" > "${OUT}.tmp"
  fi

  mv "${OUT}.tmp" "$OUT"
}

is_already_exists_response() {
  local status="$1"
  local body_file="$2"

  [[ "$status" == "400" || "$status" == "409" ]] || return 1
  grep -Eiq 'bestaat al|already exists|duplicate|unique constraint' "$body_file"
}

extract_next_link() {
  local headers_file="$1"

  tr -d '\r' < "$headers_file" \
    | awk 'BEGIN{IGNORECASE=1} /^link:/ {sub(/^[^:]+:[[:space:]]*/, "", $0); print; exit}' \
    | tr ',' '\n' \
    | awk '/rel="next"/ {if (match($0, /<[^>]+>/)) print substr($0, RSTART + 1, RLENGTH - 2)}'
}

fetch_source_or_die() {
  local url="$1"
  local body_file="$2"
  local headers_file="$3"
  local context="$4"
  local http_code

  http_code="$(request_source_once "$url" "$body_file" "$headers_file")"
  if [[ "$http_code" == "401" ]]; then
    echo
    echo "401 ontvangen van bron voor ${context}."
    prompt_source_auth
    http_code="$(request_source_once "$url" "$body_file" "$headers_file")"
    if [[ "$http_code" == "401" ]]; then
      echo "Nog steeds 401 van de bron na opnieuw invoeren van credentials. Stop." >&2
      exit 1
    fi
  fi

  if [[ "$http_code" != "200" ]]; then
    echo "GET ${url} gaf status ${http_code}. Stop." >&2
    cat "$body_file" >&2
    exit 1
  fi
}

register_missing_field() {
  local phase="$1"
  local descriptor="$2"
  local item="$3"
  local message="$4"
  local body_file

  body_file="$(mktemp)"
  printf '%s\n' "$message" > "$body_file"
  append_error "$phase" "$descriptor" "$item" "0" "$body_file"
  rm -f "$body_file"
}

post_target() {
  local path="$1"
  local phase="$2"
  local descriptor="$3"
  local payload="$4"
  local http_code
  local body_file

  body_file="$(mktemp)"
  http_code="$(request_target_once "${TARGET_BASE_URL}${path}" "$payload" "$body_file")"

  if [[ "$http_code" == "401" ]]; then
    echo
    echo "401 ontvangen van doel voor ${descriptor}."
    prompt_target_auth
    http_code="$(request_target_once "${TARGET_BASE_URL}${path}" "$payload" "$body_file")"
    if [[ "$http_code" == "401" ]]; then
      echo "Nog steeds 401 van het doel na opnieuw invoeren van credentials. Stop." >&2
      rm -f "$body_file"
      exit 1
    fi
  fi

  if [[ "$http_code" == "201" ]]; then
    if [[ "$phase" == "organisation" ]]; then
      org_success=$((org_success + 1))
    else
      git_org_success=$((git_org_success + 1))
    fi
  elif is_already_exists_response "$http_code" "$body_file"; then
    if [[ "$phase" == "organisation" ]]; then
      org_skipped=$((org_skipped + 1))
    else
      git_org_skipped=$((git_org_skipped + 1))
    fi
    echo "SKIP [${phase}] ${descriptor} -> ${http_code} (bestaat al)"
    rm -f "$body_file"
    return 0
  else
    append_error "$phase" "$descriptor" "$payload" "$http_code" "$body_file"
    if [[ "$phase" == "organisation" ]]; then
      org_failed=$((org_failed + 1))
    else
      git_org_failed=$((git_org_failed + 1))
    fi
  fi

  echo "POST [${phase}] ${descriptor} -> ${http_code}"
  rm -f "$body_file"

  if [[ "$SLEEP_SECONDS" != "0" ]]; then
    sleep "$SLEEP_SECONDS"
  fi
}

sync_organisations() {
  local next_url="${SOURCE_BASE_URL}/organisations?page=1&perPage=${PER_PAGE}"

  while [[ -n "$next_url" ]]; do
    local body_file
    local headers_file

    body_file="$(mktemp)"
    headers_file="$(mktemp)"

    fetch_source_or_die "$next_url" "$body_file" "$headers_file" "organisations"
    echo "Organisations opgehaald uit bron: ${next_url}"

    while read -r item; do
      local payload
      local uri
      local label

      uri="$(jq -r '.uri // empty' <<<"$item")"
      label="$(jq -r '.label // empty' <<<"$item")"

      if [[ -z "$uri" || -z "$label" ]]; then
        register_missing_field "organisation-build" "${uri:-missing-uri}" "$item" "ontbrekende uri of label in bronresponse"
        org_failed=$((org_failed + 1))
        echo "SKIP [organisation] ontbrekende uri of label"
        continue
      fi

      payload="$(jq -c '{uri, label}' <<<"$item")"
      post_target "/organisations" "organisation" "$uri" "$payload"
    done < <(jq -c '.[]' "$body_file")

    next_url="$(extract_next_link "$headers_file")"
    rm -f "$body_file" "$headers_file"
  done
}

sync_git_organisations() {
  local next_url="${SOURCE_BASE_URL}/git-organisations?page=1&perPage=${PER_PAGE}"

  while [[ -n "$next_url" ]]; do
    local body_file
    local headers_file

    body_file="$(mktemp)"
    headers_file="$(mktemp)"

    fetch_source_or_die "$next_url" "$body_file" "$headers_file" "git-organisations"
    echo "Git-organisations opgehaald uit bron: ${next_url}"

    while read -r item; do
      local payload
      local url
      local organisation_uri

      url="$(jq -r '.url // empty' <<<"$item")"
      organisation_uri="$(jq -r '.organisation.uri // empty' <<<"$item")"

      if [[ -z "$url" || -z "$organisation_uri" ]]; then
        register_missing_field "git-organisation-build" "${url:-missing-url}" "$item" "ontbrekende url of organisation.uri in bronresponse"
        git_org_failed=$((git_org_failed + 1))
        echo "SKIP [git-organisation] ontbrekende url of organisation.uri"
        continue
      fi

      payload="$(jq -nc --arg url "$url" --arg organisationUri "$organisation_uri" '{url:$url, organisationUri:$organisationUri}')"
      post_target "/git-organisations" "git-organisation" "$url" "$payload"
    done < <(jq -c '.[]' "$body_file")

    next_url="$(extract_next_link "$headers_file")"
    rm -f "$body_file" "$headers_file"
  done
}

require_cmd curl
require_cmd jq

if [[ $# -gt 2 ]]; then
  usage >&2
  exit 1
fi

if [[ $# -ge 1 ]]; then
  SOURCE_BASE_URL="$1"
fi
if [[ $# -ge 2 ]]; then
  TARGET_BASE_URL="$2"
fi

SOURCE_BASE_URL="${SOURCE_BASE_URL%/}"
TARGET_BASE_URL="${TARGET_BASE_URL%/}"

if [[ -z "$SOURCE_BASE_URL" || -z "$TARGET_BASE_URL" ]]; then
  usage >&2
  exit 1
fi

echo '[]' > "$OUT"

echo "Bron: ${SOURCE_BASE_URL}"
echo "Doel: ${TARGET_BASE_URL}"
echo "Errors worden opgeslagen in: ${OUT}"

if [[ "$SKIP_ORGANISATIONS" == "1" || "$SKIP_ORGANISATIONS" == "true" ]]; then
  echo "Organisations overslaan (SKIP_ORGANISATIONS=${SKIP_ORGANISATIONS})."
else
  sync_organisations
fi

if [[ "$SKIP_GIT_ORGANISATIONS" == "1" || "$SKIP_GIT_ORGANISATIONS" == "true" ]]; then
  echo "Git-organisations overslaan (SKIP_GIT_ORGANISATIONS=${SKIP_GIT_ORGANISATIONS})."
else
  sync_git_organisations
fi

echo
echo "Klaar."
echo "Organisations: succes=${org_success}, skipped=${org_skipped}, fouten=${org_failed}"
echo "Git-organisations: succes=${git_org_success}, skipped=${git_org_skipped}, fouten=${git_org_failed}"
echo "Foutenbestand: ${OUT}"
