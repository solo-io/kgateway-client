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

require_command curl
require_command jq

require_env SLACK_WEBHOOK_URL
require_env SLACK_MESSAGE

jq -n --arg text "${SLACK_MESSAGE}" '{text: $text}' | \
	curl --fail --show-error --silent \
		-X POST \
		-H "Content-Type: application/json" \
		--data @- \
		"${SLACK_WEBHOOK_URL}" >/dev/null
