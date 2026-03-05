const { execFileSync } = require("node:child_process");

function runGit(args) {
  return execFileSync("git", args, { encoding: "utf8" }).trim();
}

function parseVersion(version) {
  const match = /^v?(\d+)\.(\d+)\.(\d+)(?:-.+)?$/.exec(version || "");
  if (!match) return null;
  return {
    major: Number(match[1]),
    minor: Number(match[2]),
    patch: Number(match[3]),
  };
}

function compareVersions(a, b) {
  if (a.major !== b.major) return a.major - b.major;
  if (a.minor !== b.minor) return a.minor - b.minor;
  return a.patch - b.patch;
}

function findLatestStableTagOnMain() {
  const tags = runGit([
    "tag",
    "--merged",
    "origin/main",
    "--list",
    "v[0-9]*.[0-9]*.[0-9]*",
  ])
    .split("\n")
    .map((tag) => tag.trim())
    .filter((tag) => /^v\d+\.\d+\.\d+$/.test(tag));

  if (!tags.length) return null;

  return tags.sort((left, right) => {
    const a = parseVersion(left);
    const b = parseVersion(right);
    if (!a || !b) return 0;
    return compareVersions(a, b);
  })[tags.length - 1];
}

module.exports = {
  async analyzeCommits(_, context) {
    const mod = await import("@semantic-release/commit-analyzer");
    const analyze =
      mod?.default?.analyzeCommits ||
      mod?.analyzeCommits ||
      mod?.default ||
      mod;
    if (typeof analyze !== "function") {
      throw new TypeError(
        "Unable to load @semantic-release/commit-analyzer analyze function"
      );
    }
    const branchName = context.branch?.name || process.env.GITHUB_REF_NAME || "";
    const analyzedReleaseType = await analyze(
      { preset: "conventionalcommits" },
      context
    );

    if (branchName !== "dev") {
      return analyzedReleaseType;
    }

    // Keep dev aligned to main+1 minor, but only release when conventional commits require it.
    if (!analyzedReleaseType) {
      context.logger.log(
        "No conventional-commit release detected for dev; skipping release."
      );
      return null;
    }

    try {
      runGit(["fetch", "--no-tags", "origin", "main"]);
      runGit(["fetch", "--tags", "origin"]);
    } catch {
      // Continue with available refs if fetch is unavailable.
    }

    const latestMainTag = findLatestStableTagOnMain();
    if (!latestMainTag) {
      context.logger.log(
        "No stable main tag found; defaulting dev release type to patch."
      );
      return analyzedReleaseType;
    }

    const mainVersion = parseVersion(latestMainTag);
    if (!mainVersion) return analyzedReleaseType;

    const desiredBase = {
      major: mainVersion.major,
      minor: mainVersion.minor + 1,
      patch: 0,
    };

    const current = parseVersion(context.lastRelease?.version || "0.0.0");
    if (!current) return analyzedReleaseType;

    context.logger.log(
      "dev version policy: latest main %s -> target prerelease base %d.%d.%d",
      latestMainTag,
      desiredBase.major,
      desiredBase.minor,
      desiredBase.patch
    );

    if (compareVersions(current, desiredBase) < 0) {
      if (desiredBase.major > current.major) return "major";
      if (desiredBase.minor > current.minor) return "minor";
      return "patch";
    }

    return "patch";
  },
};
