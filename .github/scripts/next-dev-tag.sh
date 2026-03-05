#!/usr/bin/env bash
set -euo pipefail

latest_stable_tag="$(git tag -l 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1)"
if [[ -z "${latest_stable_tag}" ]]; then
  echo "unable to determine latest stable tag" >&2
  exit 1
fi

if [[ ! "${latest_stable_tag}" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "latest stable tag has unexpected format: ${latest_stable_tag}" >&2
  exit 1
fi

major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"

next_minor="$((minor + 1))"
base_version="v${major}.${next_minor}.0"

dev_numbers="$({ git tag -l "${base_version}-dev.*" || true; } | sed -E 's/^.*-dev\.([0-9]+)$/\1/' | grep -E '^[0-9]+$' || true)"

next_dev_number="1"
if [[ -n "${dev_numbers}" ]]; then
  highest_dev_number="$(echo "${dev_numbers}" | sort -n | tail -n 1)"
  next_dev_number="$((highest_dev_number + 1))"
fi

echo "${base_version}-dev.${next_dev_number}"
