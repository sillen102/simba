Release automation plan

- Goal: automate releases for root, websocket, and telemetry modules in this monorepo.
- Behavior:
  - On push/merge to `dev`: create a prerelease tag in the exact format `v<MAJOR>.<MINOR>-devN` (increment N).
  - On push/merge to `main`: compute semantic next stable version `v<MAJOR>.<MINOR>.<PATCH>` (using conventional commits) and create a prerelease for it.
  - For every run, update `websocket` and `telemetry` go.mod to depend on the new root tag, commit that change, then create three annotated git tags and three GitHub Releases (root, `websocket/`, `telemetry/`) that all point to the same commit.

- Implementation artifacts added:
  - `.releaserc.json` — semantic-release analysis config
  - `package.json` — devDependencies used by CI
  - `.github/workflows/release.yml` — GitHub Action that runs on `dev` and `main`
  - `scripts/get-latest-stable.js` — finds latest non-`-dev` tag
  - `scripts/compute-next-stable.js` — semantic analysis for `main` to compute next stable
  - `scripts/compute-dev-tag.js` — compute `vM.N-devK` for `dev` branch
  - `scripts/update-submodules.sh` — run `go get` + `go mod tidy` in submodules and commit
  - `scripts/create-tags-and-releases.sh` — create annotated tags and GitHub prereleases

Notes:
- All releases are created as prereleases per project preference.
- The workflow auto-commits submodule bumps to the triggering branch.
- Scripts are intentionally conservative and will abort on tag collision; rerun will recompute.

Testing steps (recommended):
1. Add these files in a branch or fork.
2. Run the JS scripts locally to verify computed versions:
   - `node scripts/get-latest-stable.js`
   - `node scripts/compute-dev-tag.js 0.29.1`
   - `node scripts/compute-next-stable.js`
3. Push to `dev` on a test fork to observe tag creation and commits.

If anything needs adjustment (PRs instead of direct commits, retry logic, or changelog injection), update the scripts/workflow accordingly.
