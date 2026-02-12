# GitHub Workflows Overview

This document explains the complete GitHub Actions automation system for the Twine project.

## Summary

The Twine project uses three main GitHub Actions workflows to automate the development and release process:

1. **Auto PR** (`auto-pr.yml`) - Automatically creates pull requests when the owner pushes feature branches
2. **CI** (`ci.yml`) - Runs tests on all pull requests to validate changes before merging
3. **Release** (`release.yml`) - Manages semantic versioning and releases using release-please

These workflows interact to provide a seamless development experience:
- Push a feature branch → Auto-PR creates a PR
- PR created → CI runs tests
- Tests pass → PR auto-merges
- PR merges to main → Release-please updates version and creates releases

## Feature Development Workflow

### 1. Creating Feature Branches

Follow standard Git workflow with conventional commits:

```bash
# Create and switch to a feature branch
git checkout -b feat/add-new-feature

# Make your changes
vim pkg/router/router.go

# Commit using conventional commits format
git commit -m "feat: add route parameter validation"

# Push the branch
git push -u origin feat/add-new-feature
```

### 2. Conventional Commits Format

All commits must follow the conventional commits specification for release-please to work correctly:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat:` - New feature (bumps minor version: 0.1.0 → 0.2.0)
- `fix:` - Bug fix (bumps patch version: 0.1.0 → 0.1.1)
- `feat!:` or `fix!:` - Breaking change (bumps major version: 0.1.0 → 1.0.0)
- `docs:` - Documentation only (no version bump)
- `test:` - Test changes (no version bump)
- `refactor:` - Code refactoring (no version bump)
- `chore:` - Maintenance tasks (no version bump)

**Examples:**
```bash
feat: add file-based routing support
fix: resolve infinite loop in dev watcher
feat!: change import paths to require pkg/ prefix
docs: update workflows documentation
test: add coverage for route generation
refactor: simplify router initialization
chore: update dependencies
```

### 3. Auto-PR Creation

When you (the repository owner) push a feature branch, the auto-PR workflow automatically:

1. Detects the push (only for `github.actor == 'cstone'`)
2. Checks if a PR already exists for the branch
3. If no PR exists, creates one with:
   - Title: Latest commit message
   - Body: Auto-generated description
   - Base: `main`
   - Head: Your feature branch
4. Enables auto-merge on the PR

**Important:** This only works for the repository owner. Other contributors must manually create PRs.

### 4. CI Validation

Once a PR is created, the CI workflow:

1. Checks if the PR is from release-please (skips if true)
2. Detects changed files to optimize execution
3. Skips CI if only docs/markdown files changed
4. If relevant files changed:
   - Sets up Go 1.25
   - Installs `just` command runner
   - Runs full test suite (`just test`)
   - Builds CLI to verify compilation (`just build-cli`)

### 5. Auto-Merge

After CI passes:

1. GitHub's auto-merge feature automatically merges the PR
2. The PR is merged using merge commits (not squash or rebase)
3. Linear history is maintained by branch protection rules
4. Feature branch is automatically deleted after merge

## Branch Protection Rules

The `main` branch has the following protection rules configured:

### Requirements

- ✅ **Require pull request before merging**
  - Required approvals: 0 (you're the sole owner)
  - Dismiss stale approvals when new commits are pushed
- ✅ **Require status checks to pass before merging**
  - Required check: `test` (from ci.yml)
  - Branches must be up to date before merging
- ✅ **Require linear history**
  - No merge commits from local git
  - Forces rebase or squash workflows
- ✅ **Allow specified actors to bypass required pull requests**
  - GitHub Actions bot can merge without manual approval

### Bypass Configuration

The following actors can bypass branch protection:

- `github-actions[bot]` - For automated merges
- Release-please bot - For version bump commits

This allows auto-merge to work while still protecting main from direct pushes.

### Repository Settings

Additional settings for smooth automation:

- **Automatically delete head branches** - Cleans up after merge
- **Allow auto-merge** - Enables PR auto-merge feature

## Secrets Management

### Required Secret: PAT

The workflows require a Personal Access Token (PAT) stored as a repository secret named `PAT`.

#### Why PAT is Needed

The built-in `GITHUB_TOKEN` has a limitation: workflows it triggers don't trigger other workflows. This is a security feature to prevent infinite workflow loops.

Using a PAT allows:
- `auto-pr.yml` to trigger `ci.yml` when creating PRs
- `release.yml` to trigger other workflows when merging release PRs
- Composite actions to enable auto-merge on PRs

#### Creating the PAT

1. Go to GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Click "Generate new token"
3. Configure the token:
   - **Token name:** `Twine Automation`
   - **Expiration:** 1 year (or custom)
   - **Repository access:** Only select repositories → `twine`
   - **Permissions:**
     - Contents: Read and write
     - Pull requests: Read and write
     - Metadata: Read-only (automatic)
4. Click "Generate token"
5. Copy the token (you won't see it again)

#### Adding the Secret

1. Go to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `PAT`
4. Secret: Paste the token value
5. Click "Add secret"

#### Security Considerations

- Never commit the PAT to the repository
- Use fine-grained tokens with minimal permissions
- Regularly rotate the token (before expiration)
- Monitor workflow runs for unexpected behavior
- Revoke the token immediately if compromised

## Workflows

### 1. CI Workflow (`ci.yml`)

**Purpose:** Validate all pull requests before merging

**Trigger:**
```yaml
on:
  pull_request:
    branches:
      - main
```

**What it does:**

1. Skips release-please PRs (they only update version files, no code to test)
2. Detects changed files to optimize execution
3. Skips expensive operations if only docs changed
4. Runs full test suite to catch regressions
5. Builds CLI to ensure compilation succeeds

**How it works:**

```yaml
jobs:
  test:
    if: "!startsWith(github.head_ref, 'release-please--')"
    runs-on: ubuntu-latest
    steps:
      - name: Check for relevant changes
        uses: tj-actions/changed-files@v46
        with:
          files_ignore: |
            **/*.md
            .github/**
            LICENSE
            .gitignore

      - name: Run tests (if relevant files changed)
        if: steps.changes.outputs.any_changed == 'true'
        run: just test
```

**Optimization:** The workflow uses conditional execution on every step. If only markdown files changed, the workflow passes immediately without running tests or setting up Go.

**Status check:** Branch protection requires the `test` job to pass before merging.

### 2. Auto PR Workflow (`auto-pr.yml`)

**Purpose:** Automatically create pull requests when the owner pushes feature branches

**Trigger:**
```yaml
on:
  push:
    branches-ignore:
      - main
```

**What it does:**

1. Detects when you push a feature branch
2. Checks if you're the repository owner (`github.actor == 'cstone'`)
3. Searches for existing PRs for the branch
4. Creates a new PR if none exists
5. Enables auto-merge on the PR

**Author Restriction Logic:**

```yaml
jobs:
  create-pr:
    if: github.actor == 'cstone'
```

This ensures that only your pushes trigger auto-PR creation. Other contributors must manually create PRs through the GitHub UI or CLI.

**PR Creation:**

```bash
pr_url=$(gh pr create \
  --base main \
  --head "${GITHUB_REF_NAME}" \
  --title "$(git log -1 --pretty=%s)" \
  --body "Auto-created PR from branch \`${GITHUB_REF_NAME}\`")
```

The PR title is taken from the latest commit message, so use descriptive conventional commit messages.

**Auto-Merge:**

After creating the PR, the workflow calls the `auto-merge` composite action to enable GitHub's auto-merge feature. The PR will automatically merge once CI passes.

**Composite Actions:**

The workflow uses two custom composite actions:
- `latest-pr` - Finds existing PRs for a branch
- `auto-merge` - Enables auto-merge on a PR

### 3. Release Workflow (`release.yml`)

**Purpose:** Automate semantic versioning and releases using release-please

**Trigger:**
```yaml
on:
  push:
    branches:
      - main
```

**What it does:**

The workflow has two modes of operation:

#### Mode 1: Creating Release PRs

When conventional commits are pushed to `main`:

1. Release-please analyzes commits since last release
2. Determines version bump based on commit types:
   - `feat:` → minor bump (0.1.0 → 0.2.0)
   - `fix:` → patch bump (0.1.0 → 0.1.1)
   - `feat!:` → major bump (0.1.0 → 1.0.0)
3. Creates (or updates) a release PR with:
   - Updated `CHANGELOG.md`
   - Updated version in `cmd/twine/version.go`
   - Updated `.release-please-manifest.json`
4. Enables auto-merge on the release PR

#### Mode 2: Publishing Releases

When a release PR is merged:

1. Release-please creates a GitHub release
2. Tags the release (e.g., `v0.2.0`)
3. The workflow then:
   - Sets up Go and just
   - Builds binaries for all platforms (`just build-cli-all`)
   - Uploads binaries as release assets

**Release-Please Configuration:**

```json
{
  "packages": {
    ".": {
      "release-type": "go",
      "package-name": "twine",
      "changelog-path": "CHANGELOG.md",
      "include-v-in-tag": true,
      "extra-files": [
        "cmd/twine/version.go"
      ]
    }
  }
}
```

**Version Tracking:**

The `.release-please-manifest.json` file tracks the current version:

```json
{
  ".": "0.1.2"
}
```

This file is automatically updated by release-please on each release.

**Binary Publishing:**

The workflow builds binaries for multiple platforms and uploads them to the GitHub release:

```bash
for binary in dist/*; do
  gh release upload ${{ steps.release.outputs.tag_name }} "$binary"
done
```

Users can then download platform-specific binaries from the release page.

## Composite Actions

Composite actions are reusable workflow components that encapsulate common operations.

### auto-merge (`auto-merge/action.yml`)

**Purpose:** Enable GitHub's auto-merge feature on a pull request

**Inputs:**
- `pr-url` - The PR URL to enable auto-merge on
- `token` - GitHub Personal Access Token with repo scope

**Implementation:**

```bash
gh pr merge --auto --merge "${{ inputs.pr-url }}"
```

**Merge strategy:** Uses merge commits (not squash or rebase) to preserve commit history and ensure linear history through branch protection rules.

**Error handling:** Validates PR URL before attempting merge and provides clear feedback in logs.

### latest-pr (`latest-pr/action.yml`)

**Purpose:** Find the most recent open PR for a given branch

**Inputs:**
- `branch` - Branch name to search (optional, defaults to current branch)

**Outputs:**
- `pr-url` - The URL of the latest open PR (empty string if none found)

**Implementation:**

```bash
pr_url=$(gh pr list \
  --head "$branch" \
  --base main \
  --state open \
  --json url \
  --jq '.[0].url // empty')
```

**Key features:**
- Returns empty string (not error) if no PR found
- Uses `--jq '.[0].url // empty'` for safe null handling
- Uses built-in `github.token` for read-only operations

## Workflow Interactions

Here's how the workflows interact in a typical development cycle:

### Scenario 1: Feature Development

```
1. You: git push origin feat/new-feature
   ↓
2. auto-pr.yml: Creates PR "feat: add new feature"
   ↓
3. auto-pr.yml: Enables auto-merge on PR
   ↓
4. ci.yml: Triggered by PR creation
   ↓
5. ci.yml: Runs tests and build
   ↓
6. GitHub: Auto-merges PR (tests passed)
   ↓
7. release.yml: Triggered by merge to main
   ↓
8. release.yml: Creates/updates release PR
```

### Scenario 2: Release

```
1. release.yml: Creates release PR "chore(main): release 0.2.0"
   ↓
2. ci.yml: Skipped (release-please PR)
   ↓
3. release.yml: Enables auto-merge
   ↓
4. GitHub: Auto-merges release PR (no CI required)
   ↓
5. release.yml: Creates GitHub release v0.2.0
   ↓
6. release.yml: Builds and uploads binaries
```

### Scenario 3: Documentation Update

```
1. You: git push origin docs/update
   ↓
2. auto-pr.yml: Creates PR "docs: update workflows"
   ↓
3. auto-pr.yml: Enables auto-merge
   ↓
4. ci.yml: Detects only .md files changed
   ↓
5. ci.yml: Skips tests (passes immediately)
   ↓
6. GitHub: Auto-merges PR (CI passed)
   ↓
7. release.yml: Skips release (no version bump for docs)
```

## Troubleshooting

### Common Issues

#### Issue: Auto-PR not created

**Symptoms:** Pushed a feature branch but no PR was created

**Causes:**
1. You're not the repository owner (`github.actor != 'cstone'`)
2. A PR already exists for the branch
3. The workflow failed to run

**Solutions:**
1. Check the Actions tab for workflow runs
2. Manually create PR: `gh pr create --base main`
3. Check if PAT is valid and has correct permissions

#### Issue: CI workflow not running

**Symptoms:** PR created but no CI checks appear

**Causes:**
1. PR is from release-please (intentionally skipped)
2. Only non-code files changed (intentionally skipped)
3. Workflow file has syntax errors

**Solutions:**
1. Check workflow file syntax: `actionlint .github/workflows/ci.yml`
2. Check Actions tab for error messages
3. Verify branch protection requires the `test` check

#### Issue: Auto-merge not working

**Symptoms:** PR stays open after CI passes

**Causes:**
1. Auto-merge not enabled on PR
2. Branch protection blocking merge
3. Required status checks not passing
4. Branch not up to date with base

**Solutions:**
1. Manually enable: `gh pr merge --auto --merge <PR_URL>`
2. Check branch protection settings
3. Verify all required checks pass
4. Update branch: `gh pr update <PR_NUMBER> --merge`

#### Issue: Release-please not creating release PR

**Symptoms:** Commits merged to main but no release PR created

**Causes:**
1. Commits don't follow conventional commits format
2. Commits are docs/chore only (no version bump)
3. Release PR already exists
4. PAT expired or invalid

**Solutions:**
1. Check commit messages follow `type: description` format
2. Include at least one `feat:` or `fix:` commit
3. Check for existing release PR: `gh pr list --label "autorelease: pending"`
4. Verify PAT secret is valid

#### Issue: Binary upload fails

**Symptoms:** Release created but no binaries attached

**Causes:**
1. Build failed
2. `dist/` directory empty or wrong path
3. PAT doesn't have write permissions
4. Network issues

**Solutions:**
1. Check build step logs
2. Verify `just build-cli-all` creates `dist/*` files
3. Re-generate PAT with contents write permission
4. Re-run failed workflow

### Debugging Workflows

#### View workflow runs

```bash
# List recent workflow runs
gh run list

# View specific run
gh run view <RUN_ID>

# View run logs
gh run view <RUN_ID> --log
```

#### Test workflows locally

You can't run GitHub Actions locally, but you can test components:

```bash
# Test that tests pass
just test

# Test that build works
just build-cli-all

# Test PR creation
gh pr create --base main --title "test" --body "test"

# Test auto-merge
gh pr merge --auto --merge <PR_URL>
```

#### Check workflow syntax

```bash
# Install actionlint
brew install actionlint

# Validate workflow files
actionlint .github/workflows/*.yml
```

#### Simulate workflow conditions

```bash
# Check if commit is conventional
git log -1 --pretty=%s | grep -E '^(feat|fix|docs|test|refactor|chore):'

# Check changed files
git diff --name-only main...HEAD

# Check for release-please branch
git branch -r | grep release-please
```

## Best Practices

### Commit Messages

- Always use conventional commits format
- Write descriptive commit messages (they become PR titles)
- Use `!` suffix for breaking changes: `feat!: rename router package`
- Group related changes in a single commit
- Use commit body for detailed explanations

### Branch Names

- Use descriptive branch names: `feat/add-routing`, `fix/template-bug`
- Avoid generic names: `fix`, `update`, `changes`
- Include ticket numbers if applicable: `feat/TWINE-123-routing`

### Pull Requests

- Review the auto-created PR before it merges
- Add comments or context if needed
- Cancel auto-merge if you want to make changes: `gh pr merge --disable-auto <PR_URL>`
- Use draft PRs for work-in-progress: `gh pr create --draft`

### Releases

- Let release-please manage versions (don't manually edit)
- Review release PRs before they merge
- Test pre-release versions before merging
- Update CHANGELOG.md manually if auto-generation misses context

### Security

- Regularly rotate PAT (before expiration)
- Use fine-grained tokens (not classic tokens)
- Monitor workflow runs for unexpected behavior
- Review auto-created PRs before they merge
- Enable security features:
  - Dependabot alerts
  - Secret scanning
  - Code scanning (CodeQL)

## Future Enhancements

Potential improvements to the automation:

1. **Code coverage reporting**
   - Upload coverage to Codecov or Coveralls
   - Add coverage badge to README
   - Fail CI if coverage drops

2. **Integration testing**
   - Test CLI commands in realistic scenarios
   - Test file-based routing generation
   - Test template rendering

3. **Multi-version support**
   - Backport fixes to older versions
   - Maintain multiple release branches
   - Support LTS releases

4. **Performance benchmarking**
   - Run benchmarks on PRs
   - Compare against main branch
   - Fail if performance regresses

5. **Canary releases**
   - Publish pre-release versions
   - Test with canary users
   - Promote to stable after validation

6. **Automated dependency updates**
   - Use Dependabot or Renovate
   - Auto-merge minor/patch updates
   - Test major updates with human review

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Release Please Documentation](https://github.com/googleapis/release-please)
- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [GitHub CLI Documentation](https://cli.github.com/manual/)
- [Branch Protection Rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
