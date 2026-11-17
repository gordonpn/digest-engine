#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DEEPSEEK_API_KEY:-}" ]]; then
  echo '::error::Set the DEEPSEEK_API_KEY Actions secret.'
  exit 1
fi

if [[ "${DEEPSEEK_API_BASE:-}" != 'https://api.deepseek.com' ]]; then
  echo '::error::Set DEEPSEEK_API_BASE to https://api.deepseek.com.'
  exit 1
fi

if ! status=$(curl --disable --silent --show-error --output /dev/null \
  --write-out '%{http_code}' --max-time 15 \
  --header "Authorization: Bearer ${DEEPSEEK_API_KEY}" \
  "${DEEPSEEK_API_BASE}/models") || [[ "$status" != 200 ]]; then
  echo '::error::DeepSeek endpoint or credentials failed the models request.'
  exit 1
fi
