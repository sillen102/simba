#!/bin/bash
# This script ensures all modules have a commit so they all get released together

set -e

# Touch a file in each module to ensure they all have changes
touch .release-marker
touch websocket/.release-marker
touch telemetry/.release-marker

git add .release-marker websocket/.release-marker telemetry/.release-marker
git commit -m "chore: sync module versions" --allow-empty || true
