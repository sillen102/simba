#!/usr/bin/env bash
set -euo pipefail

NEW_TAG="$1"
REPO="${GITHUB_REPOSITORY:-$(git remote get-url origin | sed -E 's#.*[:/](.+)/(.+)\\.git#\1/\2#') }"
if [ -z "$REPO" ]; then
  echo "REPO could not be determined; set GITHUB_REPOSITORY env var."
  exit 1
fi

ROOT_TAG="${NEW_TAG}"
WS_TAG="websocket/${NEW_TAG}"
TL_TAG="telemetry/${NEW_TAG}"

# ensure fresh tags
git fetch --tags

for T in "${ROOT_TAG}" "${WS_TAG}" "${TL_TAG}"; do
  if git rev-parse -q --verify "refs/tags/${T}" >/dev/null; then
    echo "Tag ${T} already exists, aborting."
    exit 1
  fi
done

# Annotated tags on HEAD (HEAD should be the commit that included submodule bumps)
git tag -a "${ROOT_TAG}" -m "release ${ROOT_TAG}"
git tag -a "${WS_TAG}" -m "release ${WS_TAG}"
git tag -a "${TL_TAG}" -m "release ${TL_TAG}"

git push origin --tags

# create GitHub releases (prerelease=true)
TOKEN="${GITHUB_TOKEN:-}"
API="https://api.github.com/repos/${REPO}/releases"

create_release() {
  local tag="$1"
  printf '{"tag_name":"%s","name":"%s","body":"Release %s","prerelease":true}' "$tag" "$tag" "$tag" |
    curl -s -X POST -H "Authorization: token ${TOKEN}" -H "Content-Type: application/json" -d @- "${API}" >/dev/null
}

if command -v gh >/dev/null 2>&1; then
  gh release create "${ROOT_TAG}" --notes "Release ${ROOT_TAG}" --prerelease
  gh release create "${WS_TAG}" --notes "Release ${WS_TAG}" --prerelease
  gh release create "${TL_TAG}" --notes "Release ${TL_TAG}" --prerelease
else
  if [ -z "${TOKEN}" ]; then
    echo "GITHUB_TOKEN not set; cannot create releases via API"
    exit 1
  fi
  create_release "${ROOT_TAG}"
  create_release "${WS_TAG}"
  create_release "${TL_TAG}"
fi

echo "Created tags & prereleases: ${ROOT_TAG}, ${WS_TAG}, ${TL_TAG}"
