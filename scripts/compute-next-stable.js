// Pure ESM: load all dependencies via dynamic import. No require().

(async () => {
  // Import ESM dependencies
  const simpleGit = (await import('simple-git')).default;
  const git = simpleGit();
  const semver = (await import('semver')).default;
  const commitAnalyzer = await import('@semantic-release/commit-analyzer');
  const analyzeCommits = commitAnalyzer.analyzeCommits || commitAnalyzer.default || commitAnalyzer;
  const conventionalCommits = (await import('conventional-changelog-conventionalcommits')).default;

  try {
    await git.fetch(['--tags']);
    const tags = (await git.tags()).all;
    const prodTags = tags.filter(t => t.startsWith('v') && !t.includes('-dev')).map(t => t.replace(/^v/, ''));
    const latest = (prodTags.length === 0) ? '0.0.0' : prodTags.sort((a, b) => semver.rcompare(a, b))[0];

    const range = latest === '0.0.0' ? '' : `v${latest}..HEAD`;
    const commitsRaw = await git.raw(['log', '--pretty=format:%H%x01%s%x01%b%x01', range]);
    const commitChunks = commitsRaw.split('\n').filter(Boolean).map(line => {
      const parts = line.split('\x01');
      return { hash: parts[0], message: (parts[1] || '') + '\n\n' + (parts[2] || '') };
    });

    // call analyzeCommits with conventionalcommits preset injected directly.
    const logger = { log: (...args) => { /* no-op */ } };
    const context = { commits: commitChunks, logger, cwd: process.cwd() };
    const releaseType = await analyzeCommits({ 
      presetConfig: conventionalCommits,
      preset: undefined  // Ensure it doesn't try to auto-resolve
    }, context);
    let next = latest;
    if (releaseType) {
      next = semver.inc(latest, releaseType);
    }
    process.stdout.write(next);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();
