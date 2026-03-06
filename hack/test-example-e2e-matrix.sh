#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

require_command() {
	local cmd="$1"
	if ! command -v "${cmd}" >/dev/null 2>&1; then
		echo "Required command not found: ${cmd}"
		exit 1
	fi
}

run_with_timeout() {
	local duration="$1"
	shift

	if command -v timeout >/dev/null 2>&1; then
		timeout "${duration}" "$@"
		return $?
	fi
	if command -v gtimeout >/dev/null 2>&1; then
		gtimeout "${duration}" "$@"
		return $?
	fi
	if command -v python3 >/dev/null 2>&1; then
		python3 - "${duration}" "$@" <<'PY'
import subprocess
import sys

duration = sys.argv[1]
cmd = sys.argv[2:]

try:
    if duration[-1] in {"s", "m", "h"}:
        seconds = float(duration[:-1]) * {"s": 1, "m": 60, "h": 3600}[duration[-1]]
    else:
        seconds = float(duration)
except (IndexError, ValueError):
    print(f"Unsupported timeout value: {duration}", file=sys.stderr)
    sys.exit(1)

try:
    result = subprocess.run(cmd, timeout=seconds)
    sys.exit(result.returncode)
except subprocess.TimeoutExpired:
    sys.exit(124)
PY
		return $?
	fi

	echo "Required command not found: timeout, gtimeout, or python3" >&2
	return 1
}

for cmd in git go kubectl kind helm docker; do
	require_command "${cmd}"
done

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

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/kgateway-example-e2e-tests.XXXXXX")"
declare -a worktrees=()
declare -a clusters=()
declare -a failures=()

awk_module_version() {
	local go_mod="$1"
	local module="$2"
	awk -v module="${module}" '
		$1 == "require" && $2 == module { print $3; exit }
		$1 == module { print $2; exit }
	' "${go_mod}"
}

resolve_enterprise_chart_version() {
	local worktree="$1"
	local ref="$2"
	local source_tag
	local merged_tag

	source_tag="$(git -C "${worktree}" log -1 --pretty=%B | sed -n 's/^Source-Tag:[[:space:]]*//p' | head -n 1)"
	if [[ -n "${source_tag}" ]]; then
		echo "${source_tag#v}"
		return 0
	fi

	if [[ "${ref}" == v* ]]; then
		echo "${ref#v}"
		return 0
	fi

	merged_tag="$(git -C "${worktree}" tag --merged HEAD --sort=-version:refname | head -n 1)"
	if [[ -n "${merged_tag}" ]]; then
		echo "${merged_tag#v}"
		return 0
	fi

	return 1
}

is_go_pseudo_version() {
	local version="$1"
	[[ "${version}" =~ [0-9]{14}-[0-9a-f]{12}$ ]]
}

cleanup() {
	for cluster in "${clusters[@]}"; do
		kind delete cluster --name "${cluster}" >/dev/null 2>&1 || true
	done
	for wt in "${worktrees[@]}"; do
		git worktree remove --force "${wt}" >/dev/null 2>&1 || true
	done
	rm -rf "${tmp_dir}"
}

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

print_result() {
	local ref="$1"
	local stage="$2"
	local result="$3"
	local log="$4"
	printf "%-24s %-36s %-7s %s\n" "${ref}" "${stage}" "${result}" "${log}"
}

record_failure() {
	local ref="$1"
	local stage="$2"
	local log="$3"
	failures+=("${ref}|${stage}|${log}")
}

echo "Running e2e example suite (skipping examples/fake-client unit-test example)..."
printf "%-24s %-36s %-7s %s\n" "REF" "STAGE" "RESULT" "LOG"

for ref in "${refs[@]}"; do
	if ! resolved_ref="$(resolve_ref "${ref}")"; then
		print_result "${ref}" "resolve-ref" "MISSING" "-"
		record_failure "${ref}" "resolve-ref" "-"
		continue
	fi

	safe_ref="$(echo "${ref}" | tr '/:@' '___')"
	wt="${tmp_dir}/${safe_ref}"
	worktrees+=("${wt}")
	git worktree add --detach "${wt}" "${resolved_ref}" >/dev/null 2>&1

	go_mod="${wt}/go.mod"
	kgw_version_raw="$(awk_module_version "${go_mod}" "github.com/kgateway-dev/kgateway/v2")"
	gw_api_version="$(awk_module_version "${go_mod}" "sigs.k8s.io/gateway-api")"
	enterprise_chart_version="$(resolve_enterprise_chart_version "${wt}" "${ref}" || true)"
	upstream_chart_version="${kgw_version_raw}"

	if is_go_pseudo_version "${upstream_chart_version}"; then
		upstream_chart_version="v${enterprise_chart_version}"
	fi

	if [[ -z "${kgw_version_raw}" || -z "${gw_api_version}" || -z "${enterprise_chart_version}" || -z "${upstream_chart_version}" ]]; then
		log="${tmp_dir}/${safe_ref}__versions.log"
		{
			echo "Failed to resolve required module versions from ${go_mod}."
			echo "github.com/kgateway-dev/kgateway/v2=${kgw_version_raw:-<empty>}"
			echo "upstream chart version=${upstream_chart_version:-<empty>}"
			echo "sigs.k8s.io/gateway-api=${gw_api_version:-<empty>}"
			echo "enterprise chart version=${enterprise_chart_version:-<empty>}"
		} >"${log}"
		print_result "${ref}" "resolve-versions" "FAIL" "${log}"
		record_failure "${ref}" "resolve-versions" "${log}"
		continue
	fi

	cluster_name="kgw-e2e-${safe_ref,,}"
	cluster_name="${cluster_name:0:48}"
	kubeconfig_path="${tmp_dir}/${safe_ref}.kubeconfig"

	if kind get clusters | grep -qx "${cluster_name}"; then
		kind delete cluster --name "${cluster_name}" >/dev/null 2>&1 || true
	fi

	clusters+=("${cluster_name}")

	log="${tmp_dir}/${safe_ref}__kind-create.log"
	if kind create cluster --name "${cluster_name}" --wait 180s >"${log}" 2>&1; then
		print_result "${ref}" "kind-create" "PASS" "${log}"
	else
		print_result "${ref}" "kind-create" "FAIL" "${log}"
		record_failure "${ref}" "kind-create" "${log}"
		kind delete cluster --name "${cluster_name}" >/dev/null 2>&1 || true
		continue
	fi

	kind get kubeconfig --name "${cluster_name}" >"${kubeconfig_path}"

	log="${tmp_dir}/${safe_ref}__install-crds.log"
	if (
		set -euo pipefail
		echo "Pinned upstream kgateway module version: ${kgw_version_raw}"
		echo "Using upstream kgateway CRD chart version: ${upstream_chart_version}"
		echo "Using enterprise CRD chart version: ${enterprise_chart_version}"
		echo "Using Gateway API version: ${gw_api_version}"

		helm upgrade --install kgateway-crds \
			oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds \
			--version "${upstream_chart_version}" \
			--namespace kgateway-system \
			--create-namespace \
			--kubeconfig "${kubeconfig_path}"

		helm upgrade --install enterprise-kgateway-crds \
			oci://us-docker.pkg.dev/solo-public/enterprise-kgateway/charts/enterprise-kgateway-crds \
			--version "${enterprise_chart_version}" \
			--namespace kgateway-system \
			--create-namespace \
			--kubeconfig "${kubeconfig_path}"

		KUBECONFIG="${kubeconfig_path}" kubectl apply --server-side -f \
			"https://github.com/kubernetes-sigs/gateway-api/releases/download/${gw_api_version}/standard-install.yaml"

		KUBECONFIG="${kubeconfig_path}" kubectl wait --for=condition=Established \
			crd/enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io --timeout=180s
		KUBECONFIG="${kubeconfig_path}" kubectl wait --for=condition=Established \
			crd/trafficpolicies.gateway.kgateway.dev --timeout=180s
		KUBECONFIG="${kubeconfig_path}" kubectl wait --for=condition=Established \
			crd/gateways.gateway.networking.k8s.io --timeout=180s
		KUBECONFIG="${kubeconfig_path}" kubectl wait --for=condition=Established \
			crd/httproutes.gateway.networking.k8s.io --timeout=180s
	) >"${log}" 2>&1; then
		print_result "${ref}" "install-crds" "PASS" "${log}"
	else
		print_result "${ref}" "install-crds" "FAIL" "${log}"
		record_failure "${ref}" "install-crds" "${log}"
		kind delete cluster --name "${cluster_name}" >/dev/null 2>&1 || true
		continue
	fi

	log="${tmp_dir}/${safe_ref}__create-update-delete.log"
	if (
		cd "${wt}/examples/create-update-delete-enterprisekgatewaytrafficpolicy"
		export GOWORK=off
		go run . -kubeconfig "${kubeconfig_path}" -namespace default -step-delay 1s
	) >"${log}" 2>&1; then
		print_result "${ref}" "create-update-delete" "PASS" "${log}"
	else
		print_result "${ref}" "create-update-delete" "FAIL" "${log}"
		record_failure "${ref}" "create-update-delete" "${log}"
	fi

	log="${tmp_dir}/${safe_ref}__out-of-cluster.log"
	out_of_cluster_rc=0
	(
		cd "${wt}/examples/out-of-cluster-client-configuration"
		export GOWORK=off
		run_with_timeout 25s go run . -kubeconfig "${kubeconfig_path}" -namespace default
	) >"${log}" 2>&1 || out_of_cluster_rc=$?

	if [[ "${out_of_cluster_rc}" -ne 0 && "${out_of_cluster_rc}" -ne 124 ]]; then
		print_result "${ref}" "out-of-cluster" "FAIL" "${log}"
		record_failure "${ref}" "out-of-cluster" "${log}"
	elif ! grep -q 'Found EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy"' "${log}"; then
		print_result "${ref}" "out-of-cluster" "FAIL" "${log}"
		record_failure "${ref}" "out-of-cluster" "${log}"
	else
		print_result "${ref}" "out-of-cluster" "PASS" "${log}"
	fi

	KUBECONFIG="${kubeconfig_path}" kubectl delete \
		enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io \
		example-enterprisekgateway-traffic-policy -n default --ignore-not-found >/dev/null 2>&1 || true

	log="${tmp_dir}/${safe_ref}__in-cluster.log"
	image_tag="kgateway-in-cluster-e2e:${safe_ref,,}"
	image_tag="${image_tag//[^a-z0-9_.:-]/-}"
	pod_name="in-cluster-example"

	in_cluster_ok=true
	if (
		set -euo pipefail
		docker build -t "${image_tag}" -f "${wt}/examples/in-cluster-client-configuration/Dockerfile" "${wt}"
		kind load docker-image "${image_tag}" --name "${cluster_name}"

		KUBECONFIG="${kubeconfig_path}" kubectl apply -f - <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kgateway-client-view
  namespace: default
rules:
- apiGroups:
  - enterprisekgateway.solo.io
  resources:
  - enterprisekgatewaytrafficpolicies
  verbs:
  - get
  - list
  - watch
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: default-kgateway-client-view
  namespace: default
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kgateway-client-view
EOF

		KUBECONFIG="${kubeconfig_path}" kubectl delete pod "${pod_name}" -n default --ignore-not-found

		KUBECONFIG="${kubeconfig_path}" kubectl run "${pod_name}" \
			-n default \
			--image="${image_tag}" \
			--image-pull-policy=IfNotPresent \
			--restart=Never \
			--env="NAMESPACE=default"

		KUBECONFIG="${kubeconfig_path}" kubectl wait --for=condition=Ready "pod/${pod_name}" -n default --timeout=180s

		for _ in $(seq 1 20); do
			KUBECONFIG="${kubeconfig_path}" kubectl logs "${pod_name}" -n default >>"${log}" 2>/dev/null || true
			if grep -q 'Found EnterpriseKgatewayTrafficPolicy "example-enterprisekgateway-traffic-policy"' "${log}"; then
				exit 0
			fi
			sleep 3
		done

		KUBECONFIG="${kubeconfig_path}" kubectl describe pod "${pod_name}" -n default >>"${log}" 2>&1 || true
		exit 1
	) >"${log}" 2>&1; then
		in_cluster_ok=true
	else
		in_cluster_ok=false
	fi

	if [[ "${in_cluster_ok}" == true ]]; then
		print_result "${ref}" "in-cluster" "PASS" "${log}"
	else
		print_result "${ref}" "in-cluster" "FAIL" "${log}"
		record_failure "${ref}" "in-cluster" "${log}"
	fi

	KUBECONFIG="${kubeconfig_path}" kubectl delete pod "${pod_name}" -n default --ignore-not-found >/dev/null 2>&1 || true
	KUBECONFIG="${kubeconfig_path}" kubectl delete rolebinding default-kgateway-client-view -n default --ignore-not-found >/dev/null 2>&1 || true
	KUBECONFIG="${kubeconfig_path}" kubectl delete role kgateway-client-view -n default --ignore-not-found >/dev/null 2>&1 || true
	KUBECONFIG="${kubeconfig_path}" kubectl delete \
		enterprisekgatewaytrafficpolicies.enterprisekgateway.solo.io \
		example-enterprisekgateway-traffic-policy -n default --ignore-not-found >/dev/null 2>&1 || true

	log="${tmp_dir}/${safe_ref}__multi-client-crud.log"
	if (
		cd "${wt}/examples/multi-client-crud"
		export GOWORK=off
		go run . -kubeconfig "${kubeconfig_path}" -namespace default
	) >"${log}" 2>&1; then
		if grep -q '^Done\.$' "${log}"; then
			print_result "${ref}" "multi-client-crud" "PASS" "${log}"
		else
			print_result "${ref}" "multi-client-crud" "FAIL" "${log}"
			record_failure "${ref}" "multi-client-crud" "${log}"
		fi
	else
		print_result "${ref}" "multi-client-crud" "FAIL" "${log}"
		record_failure "${ref}" "multi-client-crud" "${log}"
	fi

	kind delete cluster --name "${cluster_name}" >/dev/null 2>&1 || true
	done

if [[ "${#failures[@]}" -eq 0 ]]; then
	echo
	echo "All example e2e checks passed."
	exit 0
fi

echo
echo "Failed example e2e checks:"
for failure in "${failures[@]}"; do
	IFS='|' read -r ref stage log <<< "${failure}"
	echo "- ${ref} / ${stage}"
	if [[ "${log}" != "-" && -f "${log}" ]]; then
		sed -n '1,120p' "${log}"
	fi
done

exit 1
