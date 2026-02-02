import simpleGit from 'simple-git';

const git = simpleGit();

/**
 * Checks if there are any conventional commits since the last tag.
 * Returns "true" or "false" to stdout.
 *
 * Conventional commit types that warrant a release:
 * - feat: new features (minor bump)
 * - fix: bug fixes (patch bump)
 * - perf: performance improvements (patch bump)
 * - BREAKING CHANGE or !: breaking changes (major bump)
 *
 * Other types (chore, docs, refactor, test, ci, build) do NOT warrant a release.
 */
(async () => {
  try {
    await git.fetch(['--tags']);
    const tags = (await git.tags()).all;

    // Get the latest tag (any tag, including -dev tags)
    let latestTag = null;
    if (tags.length > 0) {
      // Get the most recent tag by commit date
      const tagRefs = await git.raw(['tag', '--sort=-creatordate']);
      const sortedTags = tagRefs.trim().split('\n').filter(t => t);
      latestTag = sortedTags[0];
    }

    // Get commit range
    let range;
    if (latestTag) {
      range = `${latestTag}..HEAD`;
    } else {
      // No tags exist yet, check all commits
      range = 'HEAD';
    }

    // Get commit messages in the range
    const log = await git.log({ from: latestTag || undefined, to: 'HEAD' });
    const commits = log.all.map(c => c.message);

    // Check for conventional commits that warrant a release
    // Pattern: ^(feat|fix|perf)(\(.+\))?(!)?:
    // Also check for BREAKING CHANGE in commit body
    const releasePattern = /^(feat|fix|perf)(\(.+\))?(!)?:/;
    const breakingChangePattern = /BREAKING CHANGE:/;

    const hasReleaseWorthyCommit = commits.some(msg => {
      const lines = msg.split('\n');
      const subject = lines[0];
      const body = lines.slice(1).join('\n');

      // Check if subject matches release-worthy pattern
      if (releasePattern.test(subject)) {
        return true;
      }

      // Check for breaking changes in body
      if (breakingChangePattern.test(body)) {
        return true;
      }

      return false;
    });

    process.stdout.write(hasReleaseWorthyCommit ? 'true' : 'false');
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();
