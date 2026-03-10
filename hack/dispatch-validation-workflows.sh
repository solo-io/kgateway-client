#!/usr/bin/env bash

set -euo pipefail

require_command() {
	local cmd="$1"
	if ! command -v "${cmd}" >/dev/null 2>&1; then
		echo "Required command not found: ${cmd}" >&2
		exit 1
	fi
}

require_env() {
	local name="$1"
	if [[ -z "${!name:-}" ]]; then
		echo "Required environment variable is not set: ${name}" >&2
		exit 1
	fi
}

require_command gh

require_env GH_TOKEN
require_env GITHUB_REPOSITORY
require_env VALIDATION_REF

WORKFLOW_REF="${WORKFLOW_REF:-main}"
WORKFLOWS=(
	"ref-validation.yaml"
	"example-e2e-validation.yaml"
)

for workflow in "${WORKFLOWS[@]}"; do
	echo "Dispatching ${workflow} for ref ${VALIDATION_REF}"
	gh workflow run "${workflow}" \
		--repo "${GITHUB_REPOSITORY}" \
		--ref "${WORKFLOW_REF}" \
		-f "refs=${VALIDATION_REF}"
done
