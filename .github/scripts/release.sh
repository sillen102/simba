#!/usr/bin/env bash
set -euo pipefail

BRANCH="${GITHUB_REF_NAME:-$(git rev-parse --abbrev-ref HEAD)}"
STABLE_PAT='^v[0-9]+\.[0-9]+\.[0-9]+$'
DEV_PAT='^v[0-9]+\.[0-9]+\.[0-9]+(-dev\.[0-9]+)?$'

strip_v() { echo "${1#v}"; }

get_latest_stable_tag() {
  git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-version:refname \
    | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1 || true
}

get_latest_dev_tag() {
  git tag --list 'v*-dev.*' --sort=-version:refname \
    | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+-dev\.[0-9]+$' | head -1 || true
}

has_releasable_commits() {
  local since="$1"
  local range="${since:+${since}..}HEAD"
  git log "$range" --pretty=format:"%s%n%b" 2>/dev/null \
    | grep -qE "^(feat|fix|perf)(\([^)]+\))?(!:|:)|^BREAKING[ -]CHANGE:" || return 1
}

# Returns 0 (true) if semver $1 < $2
version_lt() {
  local -a a b
  IFS='.' read -ra a <<< "$1"
  IFS='.' read -ra b <<< "$2"
  for i in 0 1 2; do
    local ai="${a[$i]:-0}" bi="${b[$i]:-0}"
    (( ai < bi )) && return 0
    (( ai > bi )) && return 1
  done
  return 1
}

release_main() {
  local last_stable
  last_stable=$(get_latest_stable_tag)

  if ! has_releasable_commits "$last_stable"; then
    echo "::notice::No releasable commits on main. Skipping."
    exit 0
  fi

  local next_version
  if [ -n "$last_stable" ]; then
    next_version=$(git cliff --bumped-version --tag-pattern "$STABLE_PAT")
  else
    next_version="v0.1.0"
  fi

  echo "Releasing ${next_version} on main"

  git cliff --tag "$next_version" --tag-pattern "$STABLE_PAT" --unreleased --prepend CHANGELOG.md
  git add CHANGELOG.md
  git commit -m "chore(release): $(strip_v "$next_version") [skip ci]"
  git tag -a "$next_version" -m "Release $next_version"
  ./.github/scripts/tag-go-submodules.sh "$(strip_v "$next_version")"
  git push origin main --follow-tags

  local release_notes
  release_notes=$(git cliff --latest --strip all --tag-pattern "$STABLE_PAT")
  gh release create "$next_version" --title "$next_version" --notes "$release_notes"

  echo "Released ${next_version}"
}

release_dev() {
  git fetch --no-tags origin main || true
  git fetch --tags origin || true

  local last_stable
  last_stable=$(
    git tag --merged origin/main --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-version:refname \
      | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1 || true
  )

  local last_dev
  last_dev=$(get_latest_dev_tag)

  local since_tag="${last_dev:-$last_stable}"

  if ! has_releasable_commits "$since_tag"; then
    echo "::notice::No releasable commits on dev. Skipping."
    exit 0
  fi

  local next_version
  if [ -n "$last_stable" ]; then
    local major minor patch
    IFS='.' read -r major minor patch <<< "$(strip_v "$last_stable")"
    local base="${major}.$((minor + 1)).0"

    local dev_num=1
    if [ -n "$last_dev" ]; then
      local last_dev_base
      last_dev_base=$(strip_v "$last_dev" | sed 's/-dev\.[0-9]*$//')
      if version_lt "$last_dev_base" "$base"; then
        dev_num=1
      else
        base="$last_dev_base"
        dev_num=$(( $(echo "$last_dev" | sed 's/.*-dev\.\([0-9]*\)$/\1/') + 1 ))
      fi
    fi
    next_version="v${base}-dev.${dev_num}"
  elif [ -n "$last_dev" ]; then
    local cur_base dev_num
    cur_base=$(strip_v "$last_dev" | sed 's/-dev\.[0-9]*$//')
    dev_num=$(( $(echo "$last_dev" | sed 's/.*-dev\.\([0-9]*\)$/\1/') + 1 ))
    next_version="v${cur_base}-dev.${dev_num}"
  else
    next_version="v0.1.0-dev.1"
  fi

  echo "Releasing ${next_version} on dev"

  git cliff --tag "$next_version" --tag-pattern "$DEV_PAT" --unreleased --prepend CHANGELOG.md
  git add CHANGELOG.md
  git commit -m "chore(release): $(strip_v "$next_version") [skip ci]"
  git tag -a "$next_version" -m "Release $next_version"
  ./.github/scripts/tag-go-submodules.sh "$(strip_v "$next_version")"
  git push origin dev --follow-tags

  echo "Released ${next_version}"
}

case "$BRANCH" in
  main) release_main ;;
  dev)  release_dev  ;;
  *)
    echo "Branch '$BRANCH' is not a release branch. Skipping."
    exit 0
    ;;
esac
