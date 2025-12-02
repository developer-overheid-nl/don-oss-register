#!/usr/bin/env bash

# Posts gitOrganisations for publishers that miss a tooiUrl in publishers.json.
# Usage: API_KEY=secret ./scripts/import_missing_tooi.sh [publishers.json]

set -euo pipefail

JSON_PATH="${1:-publishers.json}"
API_BASE="${API_BASE:-http://localhost:1337}"
API_KEY="${API_KEY:-}"

if [[ -z "${API_KEY}" ]]; then
  echo "API_KEY must be set" >&2
  exit 1
fi

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
  payload="$(jq -n --arg git "${git_url}" --arg org "${org_url}" '{gitOrganisationUrl:$git, organisationUrl:$org}')"
  local status
  status="$(curl --silent --show-error --output "${tmp}" --write-out "%{http_code}" \
    --request POST \
    --url "${API_BASE}/v1/gitOrganisations" \
    --header "content-type: application/json" \
    --header "x-api-key: ${API_KEY}" \
    --data "${payload}")" || status="000"
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
