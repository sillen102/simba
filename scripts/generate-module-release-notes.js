// Generate release notes for a specific module
// Usage: node generate-module-release-notes.js <module-path>

(async () => {
  const modulePath = process.argv[2];
  if (!modulePath) {
    console.error("Usage: node generate-module-release-notes.js <module-path>");
    process.exit(1);
  }

  // Import ESM dependencies
  const simpleGit = (await import("simple-git")).default;
  const git = simpleGit();
  const releaseNotesGenerator =
    await import("@semantic-release/release-notes-generator");
  const generateNotes =
    releaseNotesGenerator.generateNotes ||
    releaseNotesGenerator.default ||
    releaseNotesGenerator;

  try {
    await git.fetch(["--tags"]);
    const tags = (await git.tags()).all;

    // Get latest module-prefixed tag
    const moduleTags = tags.filter((t) => t.startsWith(`${modulePath}/`));
    let latestTag = null;
    if (moduleTags.length > 0) {
      const tagRefs = await git.raw([
        "tag",
        "--sort=-creatordate",
        "-l",
        `${modulePath}/*`,
      ]);
      const sortedTags = tagRefs
        .trim()
        .split("\n")
        .filter((t) => t);
      latestTag = sortedTags[0];
    }

    // Get commit range for this module
    const range = latestTag ? `${latestTag}..HEAD` : "";

    // Get commits that touch this module
    const commitsRaw = await git.raw([
      "log",
      "--pretty=format:%H%x01%s%x01%b%x01%an%x01%ae%x01",
      range,
      "--",
      modulePath,
    ]);

    if (!commitsRaw.trim()) {
      process.stdout.write(`No changes in ${modulePath} since last release.`);
      return;
    }

    const commitChunks = commitsRaw
      .split("\n")
      .filter(Boolean)
      .map((line) => {
        const parts = line.split("\x01");
        return {
          hash: parts[0],
          message: (parts[1] || "") + "\n\n" + (parts[2] || ""),
          author: {
            name: parts[3] || "",
            email: parts[4] || "",
          },
        };
      });

    // Filter out auto-bump commits
    const relevantCommits = commitChunks.filter((c) => {
      const subject = c.message.split("\n")[0];
      return !subject.includes("Bump Simba dependency to");
    });

    if (relevantCommits.length === 0) {
      process.stdout.write(`No significant changes in ${modulePath}.`);
      return;
    }

    // Generate release notes
    const logger = { log: () => {} };

    // Get repository URL from git remote
    let repositoryUrl = "https://github.com/owner/repo"; // fallback
    try {
      const remotes = await git.getRemotes(true);
      if (remotes.length > 0 && remotes[0].refs.fetch) {
        repositoryUrl = remotes[0].refs.fetch.replace(/\.git$/, "");
      }
    } catch (e) {
      // Use fallback
    }

    const context = {
      commits: relevantCommits,
      logger,
      cwd: process.cwd(),
      options: {
        repositoryUrl,
      },
      lastRelease: { gitTag: latestTag || "v0.0.0" },
      nextRelease: { gitTag: "HEAD", version: "next" },
    };

    const notes = await generateNotes(
      {
        preset: "conventionalcommits",
        writerOpts: {
          commitsSort: ["subject", "scope"],
        },
        presetConfig: {
          types: [
            { type: "feat", section: "### Features" },
            { type: "fix", section: "### Bug Fixes" },
            { type: "perf", section: "### Performance Improvements" },
            { type: "revert", section: "### Reverts" },
            { type: "docs", section: "### Documentation", hidden: false },
            { type: "style", section: "### Styles", hidden: true },
            { type: "chore", section: "### Miscellaneous", hidden: true },
            {
              type: "refactor",
              section: "### Code Refactoring",
              hidden: false,
            },
            { type: "test", section: "### Tests", hidden: true },
            { type: "build", section: "### Build System", hidden: true },
            { type: "ci", section: "### CI/CD", hidden: true },
          ],
        },
      },
      context,
    );

    process.stdout.write(notes || `Changes in ${modulePath}`);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();
