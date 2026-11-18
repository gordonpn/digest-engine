#!/usr/bin/env bash
set -euo pipefail

gh api "repos/${GITHUB_REPOSITORY}/pulls/${PR_NUMBER}" \
  --jq '"eligible=" + ((.head.repo.full_name == .base.repo.full_name and .user.type != "Bot" and .user.login != "dependabot[bot]") | tostring)' \
  >> "$GITHUB_OUTPUT"
