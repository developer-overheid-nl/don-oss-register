#!/usr/bin/env bash

# Posts gitOrganisations for publishers that miss a tooiUrl in publishers.json.
# Usage: API_KEY=secret ./scripts/import_missing_tooi.sh [publishers.json]
# Supports API_KEY and/or BEARER_TOKEN (like scripts/seed_publishers.sh).

set -euo pipefail

JSON_PATH="${1:-publishers.json}"
BASE_URL="${BASE_URL:-https://localhost:1337/v1}"

if [[ ! -f "${JSON_PATH}" ]]; then
  echo "Publishers file not found: ${JSON_PATH}" >&2
  exit 1
fi

auth_headers=()
[[ -n "${API_KEY:-}" ]] && auth_headers+=(-H "X-Api-Key: ${API_KEY}")
[[ -n "${BEARER_TOKEN:-}" ]] && auth_headers+=(-H "Authorization: Bearer ${BEARER_TOKEN}")

org_url_for_id() {
  case "$1" in
    6eb9a8a6-7c84-442f-8346-a4b9ca90c7a1) echo "https://www.gpp-woo.nl" ;;
    29869c95-a74e-4603-8313-f61fd7b46c25) echo "https://www.geonovum.nl" ;;
    6cb95fc7-a334-4173-b461-068dc03d93eb) echo "https://www.ictu.nl" ;;
    0ea49086-490d-4c7d-9de3-33a990652860) echo "https://vng.nl" ;;
    *) return 1 ;;
  esac
}

post_git_org() {
  local git_url="$1"
  local org_url="$2"
  echo "POST ${git_url} -> ${org_url}"
  local tmp
  tmp="$(mktemp)"
  local payload
  payload="$(jq -n --arg url "${git_url}" --arg organisationUri "${org_url}" '{url:$url, organisationUri:$organisationUri}')"
  local status
  local curl_args=(-sS --output "${tmp}" --write-out "%{http_code}" \
    --request POST \
    --url "${BASE_URL}/git-organisations" \
    --header "content-type: application/json" \
    --data "${payload}")
  if ((${#auth_headers[@]})); then
    curl_args+=("${auth_headers[@]}")
  fi
  status="$(curl "${curl_args[@]}")" || status="000"
  if [[ "${status}" =~ ^2 ]]; then
    echo "-> ${status}: $(cat "${tmp}")"
  else
    echo "-> ${status}: $(cat "${tmp}")" >&2
  fi
  rm -f "${tmp}"
}

# Format: id|git_url per code hosting entry lacking tooiUrl
while IFS="|" read -r id git_url; do
  org_url="$(org_url_for_id "${id}")" || { echo "Skipping unknown id ${id}" >&2; continue; }
  post_git_org "${git_url}" "${org_url}"
done < <(jq -r '.data[] | select(.tooiUrl==null) | .codeHosting[]?.url as $g | "\(.id)|\($g)"' "${JSON_PATH}")
