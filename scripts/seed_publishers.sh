#!/usr/bin/env bash
set -euo pipefail

# Seed organisations en git-organisaties vanuit publishers.json.
# BASE_URL, PUBLISHERS_FILE, API_KEY en BEARER_TOKEN kunnen via env worden gezet.

BASE_URL=${BASE_URL:-https://localhost:1337/v1}
PUBLISHERS_FILE=${PUBLISHERS_FILE:-publishers.json}

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required (brew install jq)" >&2
  exit 1
fi

if [[ ! -f "${PUBLISHERS_FILE}" ]]; then
  echo "Publishers file not found: ${PUBLISHERS_FILE}" >&2
  exit 1
fi

auth_headers=()
[[ -n "${API_KEY:-}" ]] && auth_headers+=(-H "X-Api-Key: ${API_KEY}")
[[ -n "${BEARER_TOKEN:-}" ]] && auth_headers+=(-H "Authorization: Bearer ${BEARER_TOKEN}")

post_json() {
  local url=$1
  local payload=$2
  local label=$3

  local tmp http_code
  tmp=$(mktemp)
  local curl_args=(-sS -o "${tmp}" -w "%{http_code}" -X POST "${url}" -H "Content-Type: application/json" -d "${payload}")
  if ((${#auth_headers[@]})); then
    curl_args+=("${auth_headers[@]}")
  fi
  http_code=$(curl "${curl_args[@]}")

  if [[ "${http_code}" == "200" || "${http_code}" == "201" ]]; then
    echo "${label}: ok (${http_code})"
  elif [[ "${http_code}" == "409" ]]; then
    echo "${label}: already exists (409)"
  else
    echo "${label}: failed (${http_code})" >&2
    cat "${tmp}" >&2
    rm -f "${tmp}"
    exit 1
  fi
  rm -f "${tmp}"
}

while IFS= read -r row; do
  label=$(jq -r '.description // empty' <<<"${row}")
  org_uri=$(jq -r '.tooiUrl // empty' <<<"${row}")

  if [[ -z "${label}" || -z "${org_uri}" || "${org_uri}" == "null" ]]; then
    echo "Skipping entry met lege description of tooiUrl" >&2
    continue
  fi

  org_payload=$(jq -nc --arg uri "${org_uri}" --arg lbl "${label}" '{uri:$uri,label:$lbl}')
  post_json "${BASE_URL}/organisations" "${org_payload}" "organisation ${label}"
  echo "${BASE_URL}/gitOrganisations"
  jq -r '.codeHosting[]?.url // empty' <<<"${row}" | while IFS= read -r code_host; do
    [[ -z "${code_host}" ]] && continue
    git_payload=$(jq -nc --arg gitOrganisationUrl "${code_host}" --arg organisationUrl "${org_uri}" \
      '{gitOrganisationUrl:$gitOrganisationUrl, organisationUrl:$organisationUrl}')
    post_json "${BASE_URL}/gitOrganisations" "${git_payload}" "  git org ${code_host}"
  done
done < <(jq -c '.data[]' "${PUBLISHERS_FILE}")
