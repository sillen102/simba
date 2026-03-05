const refName = process.env.GITHUB_REF_NAME || "";
const isMain = refName === "main";

const plugins = [
  [
    "@semantic-release/commit-analyzer",
    {
      preset: "conventionalcommits",
    },
  ],
  [
    "@semantic-release/release-notes-generator",
    {
      preset: "conventionalcommits",
    },
  ],
  [
    "@semantic-release/changelog",
    {
      changelogFile: "CHANGELOG.md",
    },
  ],
  [
    "@semantic-release/exec",
    {
      publishCmd:
        "./.github/scripts/tag-go-submodules.sh ${nextRelease.version}",
    },
  ],
  [
    "@semantic-release/git",
    {
      assets: ["CHANGELOG.md"],
      message:
        "chore(release): ${nextRelease.version} [skip ci]\\n\\n${nextRelease.notes}",
    },
  ],
];

if (isMain) {
  plugins.push([
    "@semantic-release/github",
    {
      successComment: false,
      failComment: false,
    },
  ]);
}

module.exports = {
  branches: [
    {
      name: "main",
      prerelease: "rc",
    },
    {
      name: "dev",
      prerelease: "dev",
    },
  ],
  tagFormat: "v${version}",
  plugins,
};
