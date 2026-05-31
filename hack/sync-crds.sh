#!/usr/bin/env bash
# Sync the generated CRDs from config/crd/bases into the Helm chart, wrapping them in a
# {{- if .Values.crds.install }} guard and stamping helm.sh/resource-policy: keep so a
# `helm uninstall` never rips the CRDs (and their data) out from under you.
#
# Run after `make manifests` whenever the API changes.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src_dir="${repo_root}/config/crd/bases"
out="${repo_root}/charts/manifestation-operator/templates/crds.yaml"

{
  echo '{{- if .Values.crds.install -}}'
  for crd in "${src_dir}"/*.yaml; do
    echo '---'
    # Drop any leading document separator, then inject the keep policy right after the
    # metadata.annotations key that controller-gen always emits.
    sed -e '1{/^---$/d;}' \
        -e 's|^  annotations:$|  annotations:\n    helm.sh/resource-policy: keep|' \
        "${crd}"
  done
  echo '{{- end }}'
} > "${out}"

echo "Wrote ${out} from ${src_dir}"
