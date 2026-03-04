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

mapfile -t example_dirs < <(find examples -mindepth 1 -maxdepth 1 -type d | sort)
if [[ "${#example_dirs[@]}" -eq 0 ]]; then
	echo "No example directories found under examples/."
	exit 1
fi

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/kgateway-example-tests.XXXXXX")"
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

resolve_ref() {
	local ref="$1"

	if git rev-parse --verify --quiet "${ref}^{commit}" >/dev/null; then
		echo "${ref}"
		return 0
	fi

	if [[ "${ref}" != origin/* ]] && git rev-parse --verify --quiet "origin/${ref}^{commit}" >/dev/null; then
		echo "origin/${ref}"
		return 0
	fi

	if [[ "${ref}" != refs/remotes/origin/* ]] && git rev-parse --verify --quiet "refs/remotes/origin/${ref}^{commit}" >/dev/null; then
		echo "refs/remotes/origin/${ref}"
		return 0
	fi

	return 1
}

echo "Testing examples with go test compile/test checks..."
printf "%-24s %-44s %-7s %s\n" "REF" "EXAMPLE" "RESULT" "LOG"

for ref in "${refs[@]}"; do
	if ! resolved_ref="$(resolve_ref "${ref}")"; then
		printf "%-24s %-44s %-7s %s\n" "${ref}" "*" "MISSING" "-"
		failures+=("${ref}|*|-")
		continue
	fi

	safe_ref="$(echo "${ref}" | tr '/:@' '___')"
	wt="${tmp_dir}/${safe_ref}"
	worktrees+=("${wt}")
	git worktree add --detach "${wt}" "${resolved_ref}" >/dev/null 2>&1

	for example_dir in "${example_dirs[@]}"; do
		example_name="${example_dir#examples/}"
		safe_example="$(echo "${example_name}" | tr '/:@' '___')"
		log="${tmp_dir}/${safe_ref}__${safe_example}.log"

		if (
			cd "${wt}/${example_dir}"
			export GOWORK=off
			if find . -maxdepth 1 -type f -name '*_test.go' | grep -q .; then
				go test ./...
			else
				go test -run '^$' ./...
			fi
		) >"${log}" 2>&1; then
			printf "%-24s %-44s %-7s %s\n" "${ref}" "${example_name}" "PASS" "${log}"
		else
			printf "%-24s %-44s %-7s %s\n" "${ref}" "${example_name}" "FAIL" "${log}"
			failures+=("${ref}|${example_name}|${log}")
		fi
	done
done

if [[ "${#failures[@]}" -eq 0 ]]; then
	echo
	echo "All ref/example checks passed."
	exit 0
fi

echo
echo "Failed ref/example checks:"
for failure in "${failures[@]}"; do
	IFS='|' read -r ref example log <<< "${failure}"
	echo "- ${ref} / ${example}"
	if [[ "${log}" != "-" && -f "${log}" ]]; then
		sed -n '1,60p' "${log}"
	fi
done

exit 1
