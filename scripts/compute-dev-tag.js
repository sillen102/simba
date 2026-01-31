import simpleGit from 'simple-git';
import semver from 'semver';

const git = simpleGit();

(async () => {
  try {
    const stable = process.argv[2] || '0.0.0';
    await git.fetch(['--tags']);
    const tags = (await git.tags()).all;

    const coerced = semver.coerce(stable) || { major: 0, minor: 0 };
    const major = coerced.major;
    const minor = coerced.minor + 1;
    const base = `${major}.${minor}`;

    const devRegex = new RegExp(`^v${major}\\.${minor}-dev(\\d+)$`);
    const existingNums = tags
      .map(t => {
        const m = t.match(devRegex);
        return m ? parseInt(m[1], 10) : null;
      })
      .filter(n => n !== null)
      .sort((a,b) => a - b);

    const nextN = existingNums.length === 0 ? 1 : (existingNums[existingNums.length - 1] + 1);
    const newTag = `v${base}-dev${nextN}`;

    if (tags.includes(newTag)) {
      throw new Error('computed tag already exists: ' + newTag);
    }
    process.stdout.write(newTag);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
})();
