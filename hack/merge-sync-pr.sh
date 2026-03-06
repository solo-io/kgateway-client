#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

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
require_command jq

require_env GH_TOKEN
require_env GITHUB_REPOSITORY
require_env PR_NUMBER
require_env PR_HEAD_SHA
require_env EXPECTED_CHECKS

SELF_CHECK_NAME="${SELF_CHECK_NAME:-merge-sync-pr}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-180}"
SLEEP_SECONDS="${SLEEP_SECONDS:-20}"

read -r -a expected_checks <<< "${EXPECTED_CHECKS}"
if [[ "${#expected_checks[@]}" -eq 0 ]]; then
	echo "EXPECTED_CHECKS did not contain any check names" >&2
	exit 1
fi

checks_ready=false

for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
	pr_json="$(gh pr view "${PR_NUMBER}" --repo "${GITHUB_REPOSITORY}" --json state,isDraft,statusCheckRollup)"
	pr_state="$(jq -r '.state' <<<"${pr_json}")"
	pr_draft="$(jq -r '.isDraft' <<<"${pr_json}")"

	if [[ "${pr_state}" != "OPEN" ]]; then
		echo "PR #${PR_NUMBER} is no longer open; nothing to merge"
		exit 0
	fi
	if [[ "${pr_draft}" == "true" ]]; then
		echo "PR #${PR_NUMBER} is draft; nothing to merge"
		exit 0
	fi

	failing_checks="$(jq -r --arg self "${SELF_CHECK_NAME}" '
		.statusCheckRollup[]
		| select(.name != $self)
		| select(
				(.__typename == "CheckRun" and .status == "COMPLETED" and (.conclusion != "SUCCESS" and .conclusion != "SKIPPED" and .conclusion != "NEUTRAL"))
				or
				(.__typename == "StatusContext" and (.state == "FAILURE" or .state == "ERROR"))
			)
		| "\(.name):\(
				if .__typename == "CheckRun" then .conclusion else .state end
			)"
	' <<<"${pr_json}")"

	if [[ -n "${failing_checks}" ]]; then
		echo "Failing sync PR checks detected:"
		echo "${failing_checks}"
		exit 1
	fi

	declare -a missing_checks=()
	declare -a pending_checks=()

	for check_name in "${expected_checks[@]}"; do
		present_count="$(jq -r --arg name "${check_name}" '[.statusCheckRollup[] | select(.name == $name)] | length' <<<"${pr_json}")"
		success_count="$(jq -r --arg name "${check_name}" '
			[
				.statusCheckRollup[]
				| select(.name == $name)
				| select(
						(.__typename == "CheckRun" and .status == "COMPLETED" and (.conclusion == "SUCCESS" or .conclusion == "SKIPPED" or .conclusion == "NEUTRAL"))
						or
						(.__typename == "StatusContext" and .state == "SUCCESS")
					)
			] | length
		' <<<"${pr_json}")"
		pending_count="$(jq -r --arg name "${check_name}" '
			[
				.statusCheckRollup[]
				| select(.name == $name)
				| select(
						(.__typename == "CheckRun" and .status != "COMPLETED")
						or
						(.__typename == "StatusContext" and (.state == "PENDING" or .state == "EXPECTED"))
					)
			] | length
		' <<<"${pr_json}")"

		if [[ "${present_count}" == "0" ]]; then
			missing_checks+=("${check_name}")
		elif [[ "${success_count}" != "0" ]]; then
			continue
		elif [[ "${pending_count}" != "0" ]]; then
			pending_checks+=("${check_name}")
		else
			echo "Check ${check_name} did not reach a successful terminal state"
			exit 1
		fi
	done

	if [[ "${#missing_checks[@]}" -eq 0 && "${#pending_checks[@]}" -eq 0 ]]; then
		echo "All expected sync PR checks completed successfully"
		checks_ready=true
		break
	fi

	echo "Waiting for sync PR checks to complete (attempt ${attempt}/${MAX_ATTEMPTS})"
	if [[ "${#missing_checks[@]}" -gt 0 ]]; then
		echo "Missing checks: ${missing_checks[*]}"
	fi
	if [[ "${#pending_checks[@]}" -gt 0 ]]; then
		echo "Pending checks: ${pending_checks[*]}"
	fi
	sleep "${SLEEP_SECONDS}"
done

if [[ "${checks_ready}" != "true" ]]; then
	echo "Timed out waiting for sync PR checks to complete"
	exit 1
fi

gh pr merge "${PR_NUMBER}" \
	--repo "${GITHUB_REPOSITORY}" \
	--squash \
	--match-head-commit "${PR_HEAD_SHA}" \
	--delete-branch=false
