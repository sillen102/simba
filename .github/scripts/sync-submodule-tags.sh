#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <root-tag>" >&2
  exit 1
fi

root_tag="$1"
if [[ ! "${root_tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.]+)?$ ]]; then
  echo "root tag format is invalid: ${root_tag}" >&2
  exit 1
fi

for module in websocket telemetry; do
  module_tag="${module}/${root_tag}"

  if git rev-parse -q --verify "refs/tags/${module_tag}" >/dev/null; then
    echo "tag ${module_tag} already exists"
    continue
  fi

  git tag -a "${module_tag}" -m "release ${module_tag}"
  git push origin "refs/tags/${module_tag}"
  echo "created tag ${module_tag}"
done
