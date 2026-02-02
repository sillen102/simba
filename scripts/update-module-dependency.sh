#!/bin/bash
# Update a Go module's dependency on the root simba module
# Usage: ./update-module-dependency.sh <module-dir> <new-tag> <repo-url> <branch>

set -euo pipefail

MODULE_DIR="$1"
NEW_TAG="$2"
REPO_URL="$3"
BRANCH="$4"
REPO_NAME="$5"

cd "$MODULE_DIR"

git config user.email "actions@github.com"
git config user.name "github-actions[bot]"

echo "Updating $MODULE_DIR dependency to $REPO_NAME@$NEW_TAG"
go mod edit -require="$REPO_NAME@$NEW_TAG"

# Retry go mod tidy if it fails
go mod tidy || (echo "Retrying go mod tidy..." && sleep 5 && go mod tidy)

# Commit if there are changes
git add go.mod go.sum || true

if git diff --staged --quiet; then
  echo "No dependency changes for $MODULE_DIR"
  echo "changed=false" >> "$GITHUB_OUTPUT"
else
  git commit -m "Bump Simba dependency to ${NEW_TAG}"

  # Sync with remote
  git remote remove origin 2>/dev/null || true
  git remote add origin "$REPO_URL"
  git fetch origin "$BRANCH"

  # Try rebase, fall back to merge
  if git rev-parse --verify "origin/$BRANCH" >/dev/null 2>&1; then
    if ! git rebase "origin/$BRANCH"; then
      echo "Rebase failed, merging instead"
      git rebase --abort || true
      git merge --no-edit "origin/$BRANCH" || true
    fi
  fi

  git push "$REPO_URL" "HEAD:refs/heads/$BRANCH"
  echo "changed=true" >> "$GITHUB_OUTPUT"
fi
