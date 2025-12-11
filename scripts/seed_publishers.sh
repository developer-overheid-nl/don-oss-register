#!/usr/bin/env bash
set -euo pipefail

# Seed organisations en git-organisaties vanuit publishers.json.
# BASE_URL, PUBLISHERS_FILE, API_KEY en BEARER_TOKEN kunnen via env worden gezet.
# Combineert de eerdere import_missing_tooi functionaliteit: entries zonder tooiUrl
# worden via een bekende mapping alsnog aan een organisatie gekoppeld.

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

org_url_for_id() {
  case "$1" in
    6eb9a8a6-7c84-442f-8346-a4b9ca90c7a1) echo "https://www.gpp-woo.nl" ;;
    29869c95-a74e-4603-8313-f61fd7b46c25) echo "https://www.geonovum.nl" ;;
    6cb95fc7-a334-4173-b461-068dc03d93eb) echo "https://www.ictu.nl" ;;
    0ea49086-490d-4c7d-9de3-33a990652860) echo "https://vng.nl" ;;
    *) return 1 ;;
  esac
}

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
  id=$(jq -r '.id // empty' <<<"${row}")

  if [[ -z "${org_uri}" || "${org_uri}" == "null" ]]; then
    mapped_uri=$(org_url_for_id "${id}" || true)
    if [[ -n "${mapped_uri}" ]]; then
      org_uri="${mapped_uri}"
      echo "Gebruik mapping voor id ${id}: ${org_uri}"
    else
      echo "Skipping entry met lege description of tooiUrl (id=${id}, label=${label})" >&2
      continue
    fi
  fi

  if [[ -z "${label}" ]]; then
    echo "Skipping entry met lege description (id=${id})" >&2
    continue
  fi

  org_payload=$(jq -nc --arg uri "${org_uri}" --arg lbl "${label}" '{uri:$uri,label:$lbl}')
  post_json "${BASE_URL}/organisations" "${org_payload}" "organisation ${label}"
  jq -r '.codeHosting[]?.url // empty' <<<"${row}" | while IFS= read -r code_host; do
    [[ -z "${code_host}" ]] && continue
    git_payload=$(jq -nc --arg url "${code_host}" --arg organisationUri "${org_uri}" \
      '{url:$url, organisationUri:$organisationUri}')
    post_json "${BASE_URL}/git-organisations" "${git_payload}" "  git org ${code_host}"
  done
done < <(jq -c '.data[]' "${PUBLISHERS_FILE}")
