#!/usr/bin/env bash
# Merge docs-tfplugindocs/ and docs-plugin/ into docs/.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SDK="${ROOT}/docs-tfplugindocs"
PLUGIN="${ROOT}/docs-plugin"
OUT="${ROOT}/docs"

[[ -d "$SDK" ]] || { echo "missing docs-tfplugindocs/; run gen-docs first" >&2; exit 1; }

rm -rf "$OUT"
cp -R "$SDK" "$OUT"
for section in resources data-sources; do
	[[ -d "${PLUGIN}/${section}" ]] || continue
	mkdir -p "${OUT}/${section}"
	cp -R "${PLUGIN}/${section}/." "${OUT}/${section}/"
done
