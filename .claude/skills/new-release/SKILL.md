---
name: new-release
description: Create semantic version releases for Go projects using conventional commits. Use when user invokes /new-release or asks to "create a release", "bump version", or "prepare a new release". Analyzes commits since last release, determines version bump (major/minor/patch), updates version files, generates CHANGELOG entries, creates git tags and GitHub releases.
---

# Release Management Skill

Create intelligent, conventional-commit-based releases for the Twine project.

## Workflow

When invoked, execute the following steps:

### 1. Find Latest Release

Use git tags to find the most recent release:

```bash
git describe --tags --abbrev=0 --match "twine-v*"
```

If no tags exist, assume this is the first release (v0.1.0).

Extract version from tag (e.g., `twine-v0.3.0` → `0.3.0`).

### 2. Get Commits Since Last Release

List all commits since the last tag:

```bash
git log <last_tag>..HEAD --pretty=format:"%h|%s|%b"
```

If no last tag exists, use:

```bash
git log --pretty=format:"%h|%s|%b"
```

### 3. Analyze Commits for Version Bump

Parse commit messages using conventional commits format:

**Major Bump (X.0.0):**
- Commit subject contains `!` after type (e.g., `feat!:`, `fix!:`)
- Commit body contains `BREAKING CHANGE:` footer

**Minor Bump (0.X.0):**
- Commit type is `feat:`

**Patch Bump (0.0.X):**
- Commit type is `fix:`

**No Bump:**
- Other types: `docs:`, `chore:`, `test:`, `refactor:`, `style:`, `ci:`, `perf:`

Apply the **highest** bump type found. If no release-worthy commits exist, inform the user and exit.

### 4. Calculate New Version

Apply bump to current version:
- Current: `0.3.0`, Patch bump → `0.3.1`
- Current: `0.3.0`, Minor bump → `0.4.0`
- Current: `0.3.0`, Major bump → `1.0.0`

### 5. Update Version File

Read `cmd/twine/commands/version.go` and update the version variable:

```go
// FROM:
var version = "dev"

// TO:
var version = "0.4.0"
```

Use the Edit tool (not Write) to preserve file history.

### 6. Generate CHANGELOG Entry

Create a new changelog entry and insert it after the `# Changelog` header in `CHANGELOG.md`.

**Format:**
```markdown
## [X.Y.Z](https://github.com/cstone-io/twine/compare/twine-vA.B.C...twine-vX.Y.Z) (YYYY-MM-DD)


### Features

* commit message without feat: prefix ([hash](https://github.com/cstone-io/twine/commit/hash))

### Bug Fixes

* commit message without fix: prefix ([hash](https://github.com/cstone-io/twine/commit/hash))

### BREAKING CHANGES

* description of breaking change
```

**Sections:**
- `### Features` - for `feat:` commits
- `### Bug Fixes` - for `fix:` commits
- `### BREAKING CHANGES` - for commits with breaking changes

**Formatting:**
- Remove type prefix from commit messages (e.g., `feat: add foo` → `add foo`)
- Link commit hash to GitHub
- Use current date in ISO format (YYYY-MM-DD)
- Compare link should reference tag format `twine-vX.Y.Z`

**Important:** Read the existing CHANGELOG.md to match the exact formatting style used in previous entries.

### 7. Create Git Tag

Create an annotated tag with the Twine-specific format:

```bash
git tag -a "twine-vX.Y.Z" -m "Release X.Y.Z"
```

**Note:** Do NOT push the tag yet - user will review and push manually.

### 8. Create GitHub Release

Use the `gh` CLI to create a GitHub release:

```bash
gh release create "twine-vX.Y.Z" \
  --title "twine: vX.Y.Z" \
  --notes "<changelog_content>"
```

Where `<changelog_content>` is the markdown from the CHANGELOG entry (Features, Bug Fixes, etc.).

### 9. Present Summary

Show the user:
- ✅ Updated files: `cmd/twine/commands/version.go`, `CHANGELOG.md`
- ✅ Created tag: `twine-vX.Y.Z`
- ✅ Created GitHub release: `twine-vX.Y.Z`
- 📋 Next steps: Review changes, commit, and push to remote

## Edge Cases

### No Release Commits

If only non-release commits exist (docs, chore, test):
```
No release-worthy commits found since last release.
Only documentation, tests, or chores have been committed.
Consider manually creating a release if needed.
```

### Tag Already Exists

If the calculated tag already exists:
```
Error: Tag twine-vX.Y.Z already exists.
Please delete the tag manually if you want to recreate it:
  git tag -d twine-vX.Y.Z
  git push origin :refs/tags/twine-vX.Y.Z
```

### Dirty Working Directory

If `git status --porcelain` shows uncommitted changes:
```
Warning: You have uncommitted changes in your working directory.
The release will modify version.go and CHANGELOG.md.
```

Use AskUserQuestion to confirm whether to continue.

### Not on Main Branch

If current branch is not `main`:
```
Warning: You are not on the main branch (currently on: <branch>).
Releases are typically created from main.
```

Use AskUserQuestion to confirm whether to continue.

### First Release

If no previous tags exist:
- Start from version `0.1.0`
- Include all commits in CHANGELOG
- Compare link should point to first commit: `https://github.com/cstone-io/twine/commits/twine-v0.1.0`

## Tools to Use

- **Read** - Read `version.go`, `CHANGELOG.md`, git command output
- **Edit** - Update `version.go` (change version = "dev" to actual version)
- **Edit** - Update `CHANGELOG.md` (insert new entry after header)
- **Bash** - Execute git and gh CLI commands
- **AskUserQuestion** - Handle edge cases requiring user decisions

## Important Notes

- **NO manifest file** - Never read or create `.release-please-manifest.json`
- **Tag format** - Always use `twine-vX.Y.Z` (with "twine-" prefix)
- **Repository** - `github.com/cstone-io/twine`
- **Version location** - `cmd/twine/commands/version.go` (line 10: `version = "dev"`)
- **CHANGELOG location** - Root of repository
- **Comparison links** - Use tag format in URLs, not raw version numbers
- **Date format** - ISO 8601: YYYY-MM-DD (e.g., 2026-02-12)

## Example Output

```
🎉 Release twine-v0.4.0 created!

📝 Changes:
  - cmd/twine/commands/version.go (version updated to 0.4.0)
  - CHANGELOG.md (new release entry added)

🏷️  Git tag: twine-v0.4.0 (local only, not pushed)

📦 GitHub release: https://github.com/cstone-io/twine/releases/tag/twine-v0.4.0

📋 Next steps:
  1. Review the changes above
  2. Commit: git add -A && git commit -m "chore(main): release twine 0.4.0"
  3. Push: git push origin main
  4. Push tag: git push origin twine-v0.4.0
```

## Commit Message Parsing Examples

```
feat: add new router feature
→ Minor bump, goes in "Features" section

fix: resolve template rendering bug
→ Patch bump, goes in "Bug Fixes" section

feat!: redesign routing API
→ Major bump, goes in "Features" and "BREAKING CHANGES" sections

fix: security vulnerability in auth

BREAKING CHANGE: Auth tokens now expire after 1 hour
→ Major bump, goes in "Bug Fixes" and "BREAKING CHANGES" sections

docs: update README
→ No bump, not included in CHANGELOG

chore: bump dependencies
→ No bump, not included in CHANGELOG
```

## Version Calculation Examples

```
Current: 0.3.0
Commits: fix: bug A, fix: bug B
→ New: 0.3.1 (patch)

Current: 0.3.0
Commits: feat: feature A, fix: bug B
→ New: 0.4.0 (minor wins over patch)

Current: 0.3.0
Commits: feat!: breaking feature
→ New: 1.0.0 (major)

Current: 1.5.2
Commits: docs: update, chore: deps
→ No release (no version change)
```
