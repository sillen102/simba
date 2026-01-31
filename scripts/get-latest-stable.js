const git = require('simple-git')();

(async () => {
  try {
    await git.fetch(['--tags']);
    const tags = (await git.tags()).all;
    const prod = tags
      .filter(t => t.startsWith('v') && !t.includes('-dev'))
      .map(t => t.replace(/^v/, ''));
    if (prod.length === 0) {
      process.stdout.write('0.0.0');
      return;
    }
    const semver = require('semver');
    prod.sort((a,b) => semver.rcompare(a,b));
    process.stdout.write(prod[0]);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();
