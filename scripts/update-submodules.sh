#!/usr/bin/env bash
set -euo pipefail

NEW_TAG="$1"
MODULES=(websocket telemetry)

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

changed=false
# detect root module path from root go.mod
ROOT_MOD=$(awk '/^module/ {print $2; exit}' go.mod)
if [ -z "$ROOT_MOD" ]; then
  echo "Could not detect root module path from go.mod"
  exit 1
fi

for m in "${MODULES[@]}"; do
  if [ -d "${m}" ]; then
    pushd "${m}" >/dev/null
    echo "Updating ${m} -> ${ROOT_MOD}@${NEW_TAG}"
    set +e
    go get "${ROOT_MOD}@${NEW_TAG}"
    rc=$?
    set -e
    if [ $rc -ne 0 ]; then
      echo "go get failed for ${ROOT_MOD}@${NEW_TAG} in ${m} (exit ${rc})."
      echo "Continuing: the workflow will still attempt to commit any go.mod/go.sum changes."
    fi
    go mod tidy
    if [ -n "$(git status --porcelain)" ]; then
      git add go.mod go.sum || true
      changed=true
    fi
    popd >/dev/null
  else
    echo "Module dir ${m} not found; skipping"
  fi
done

if [ "${changed}" = true ]; then
  git commit -m "chore(release): bump submodules to ${NEW_TAG} [skip ci]" || true
  git push
else
  echo "No go.mod/go.sum changes detected."
fi
