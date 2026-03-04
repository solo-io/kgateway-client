#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

declare -a refs
if [[ $# -gt 0 ]]; then
	refs=("$@")
else
	main_ref="main"
	if ! git rev-parse --verify --quiet "${main_ref}^{commit}" >/dev/null; then
		if git rev-parse --verify --quiet "origin/main^{commit}" >/dev/null; then
			main_ref="origin/main"
		elif git rev-parse --verify --quiet "refs/remotes/origin/main^{commit}" >/dev/null; then
			main_ref="refs/remotes/origin/main"
		fi
	fi

	refs=("${main_ref}")
	while IFS= read -r tag; do
		refs+=("${tag}")
	done < <(git for-each-ref --sort=version:refname --format='%(refname:short)' refs/tags)
fi

if [[ "${#refs[@]}" -eq 0 ]]; then
	echo "No refs found to test."
	exit 1
fi

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/kgateway-ref-tests.XXXXXX")"
declare -a worktrees=()
declare -a failures=()

cleanup() {
	for wt in "${worktrees[@]}"; do
		git worktree remove --force "${wt}" >/dev/null 2>&1 || true
	done
	rm -rf "${tmp_dir}"
}
# Ensure temporary worktrees and logs are removed on all script exits.
trap cleanup EXIT

echo "Testing refs with: go test ./..."
printf "%-24s %-7s %s\n" "REF" "RESULT" "LOG"

for ref in "${refs[@]}"; do
	if ! git rev-parse --verify --quiet "${ref}^{commit}" >/dev/null; then
		printf "%-24s %-7s %s\n" "${ref}" "MISSING" "-"
		failures+=("${ref}")
		continue
	fi

	safe_ref="$(echo "${ref}" | tr '/:@' '___')"
	wt="${tmp_dir}/${safe_ref}"
	log="${tmp_dir}/${safe_ref}.log"
	worktrees+=("${wt}")

	git worktree add --detach "${wt}" "${ref}" >/dev/null 2>&1
	if (
		cd "${wt}"
		export GOWORK=off
		go test ./...
	) >"${log}" 2>&1; then
		printf "%-24s %-7s %s\n" "${ref}" "PASS" "${log}"
	else
		printf "%-24s %-7s %s\n" "${ref}" "FAIL" "${log}"
		failures+=("${ref}:${log}")
	fi
done

if [[ "${#failures[@]}" -eq 0 ]]; then
	echo
	echo "All refs passed."
	exit 0
fi

echo
echo "Failed refs:"
for failure in "${failures[@]}"; do
	ref="${failure%%:*}"
	log="${failure#*:}"
	echo "- ${ref}"
	if [[ "${log}" != "${ref}" && -f "${log}" ]]; then
		sed -n '1,60p' "${log}"
	fi
done

exit 1
