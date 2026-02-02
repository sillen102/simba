// Pure ESM: load all dependencies via dynamic import.

(async () => {
  // Import ESM dependencies
  const simpleGit = (await import("simple-git")).default;
  const git = simpleGit();
  const releaseNotesGenerator =
    await import("@semantic-release/release-notes-generator");
  const generateNotes =
    releaseNotesGenerator.generateNotes ||
    releaseNotesGenerator.default ||
    releaseNotesGenerator;
  const conventionalCommits = (
    await import("conventional-changelog-conventionalcommits")
  ).default;

  try {
    await git.fetch(["--tags"]);
    const tags = (await git.tags()).all;

    // Get the latest tag (any tag, including -dev tags)
    let latestTag = null;
    if (tags.length > 0) {
      // Get the most recent tag by commit date
      const tagRefs = await git.raw(["tag", "--sort=-creatordate"]);
      const sortedTags = tagRefs
        .trim()
        .split("\n")
        .filter((t) => t);
      latestTag = sortedTags[0];
    }

    // Get commit range
    const range = latestTag ? `${latestTag}..HEAD` : "";

    // Get commits in the range
    const commitsRaw = await git.raw([
      "log",
      "--pretty=format:%H%x01%s%x01%b%x01%an%x01%ae%x01",
      range,
    ]);

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

    if (commitChunks.length === 0) {
      process.stdout.write("No commits since last release.");
      return;
    }

    // Generate release notes using conventional commits preset
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
      commits: commitChunks,
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

    // Output the generated notes
    process.stdout.write(notes || "No significant changes.");
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();
