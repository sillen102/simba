const git = require('simple-git')();
const analyze = require('@semantic-release/commit-analyzer');

(async () => {
  try {
    await git.fetch(['--tags']);
    const tags = (await git.tags()).all;
    const prodTags = tags.filter(t => t.startsWith('v') && !t.includes('-dev')).map(t => t.replace(/^v/, ''));
    const semver = require('semver');
    const latest = (prodTags.length === 0) ? '0.0.0' : prodTags.sort((a,b) => semver.rcompare(a,b))[0];

    const range = latest === '0.0.0' ? '' : `v${latest}..HEAD`;
    const commitsRaw = await git.raw(['log', '--pretty=format:%H%x01%s%x01%b%x01', range]);
    const commitChunks = commitsRaw.split('\n').filter(Boolean).map(line => {
      const parts = line.split('\x01');
      return { hash: parts[0], message: (parts[1] || '') + '\n\n' + (parts[2] || '') };
    });

    const pluginConfig = { preset: 'conventionalcommits' };
    const releaseType = await analyze(pluginConfig, { commits: commitChunks });
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
