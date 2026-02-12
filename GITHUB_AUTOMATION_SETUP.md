# GitHub Automation Setup Checklist

This checklist guides you through setting up the complete GitHub Actions automation for Twine.

## ✅ Phase 1: Files Created

All workflow files, composite actions, and configuration have been created:

- [x] `.github/workflows/ci.yml` - CI test automation
- [x] `.github/workflows/auto-pr.yml` - Auto PR creation
- [x] `.github/workflows/release.yml` - Release automation (replaces old version)
- [x] `.github/actions/auto-merge/action.yml` - Auto-merge composite action
- [x] `.github/actions/latest-pr/action.yml` - Find PR composite action
- [x] `.release-please-config.json` - Release-please configuration
- [x] `.release-please-manifest.json` - Version tracking
- [x] `internal/docs/WORKFLOWS.md` - Complete documentation

## 🔧 Phase 2: GitHub Repository Configuration

### Step 1: Create Personal Access Token (PAT)

1. Go to: https://github.com/settings/tokens?type=beta
2. Click "Generate new token"
3. Configure the token:
   - **Token name:** `Twine Automation`
   - **Expiration:** 1 year (or custom)
   - **Repository access:** Only select repositories → `cstone-io/twine`
   - **Permissions:**
     - **Contents:** Read and write
     - **Pull requests:** Read and write
     - **Metadata:** Read-only (automatic)
4. Click "Generate token"
5. **Copy the token** (you won't see it again!)

### Step 2: Add PAT as Repository Secret

1. Go to: https://github.com/cstone-io/twine/settings/secrets/actions
2. Click "New repository secret"
3. Configure:
   - **Name:** `PAT`
   - **Secret:** Paste the token from Step 1
4. Click "Add secret"

### Step 3: Configure Branch Protection

1. Go to: https://github.com/cstone-io/twine/settings/branches
2. Click "Add rule" (or edit existing rule for `main`)
3. Configure:

   **Branch name pattern:** `main`

   **Protect matching branches:**
   - [x] Require a pull request before merging
     - Required approvals: `0` (you're the owner)
     - [x] Dismiss stale pull request approvals when new commits are pushed
   - [x] Require status checks to pass before merging
     - [x] Require branches to be up to date before merging
     - **Required checks:** Add `test` (will appear after first CI run)
   - [x] Require linear history
   - [x] Allow specified actors to bypass required pull requests
     - Add: `github-actions` (the bot account)

4. Click "Create" or "Save changes"

### Step 4: Enable Auto-Merge

1. Go to: https://github.com/cstone-io/twine/settings
2. Scroll to "Pull Requests" section
3. Check:
   - [x] Allow auto-merge
   - [x] Automatically delete head branches

## 🧪 Phase 3: Testing

### Test 1: CI Workflow

```bash
# Create a test branch
git checkout -b test/ci-automation

# Make a trivial change
echo "// CI test" >> pkg/router/router.go

# Commit with conventional format
git add .
git commit -m "test: verify CI workflow automation"

# Push to trigger auto-PR
git push -u origin test/ci-automation
```

**Expected behavior:**
1. Auto-PR workflow creates a PR
2. CI workflow runs tests
3. PR auto-merges after tests pass
4. Branch is automatically deleted

**Verify on GitHub:**
- Go to: https://github.com/cstone-io/twine/pulls
- Check that PR was created
- Check Actions tab for workflow runs
- Confirm PR merged automatically

### Test 2: Docs-Only Change (CI Skip)

```bash
# Create a docs branch
git checkout main
git pull
git checkout -b docs/test-skip

# Change only markdown
echo "## Test" >> README.md

# Commit and push
git add .
git commit -m "docs: test CI skip optimization"
git push -u origin docs/test-skip
```

**Expected behavior:**
1. Auto-PR creates PR
2. CI detects docs-only change and skips tests
3. PR auto-merges immediately

### Test 3: Release-Please

```bash
# After test PRs merge, check for release PR
gh pr list --label "autorelease: pending"
```

**Expected behavior:**
1. Release-please detects `test:` commit
2. Since `test:` doesn't bump version, no release PR created
3. Next `feat:` or `fix:` commit will trigger release PR

To trigger a release:

```bash
git checkout main
git pull
git checkout -b feat/trigger-release

# Make a feature change
echo "// Trigger release" >> pkg/router/router.go

# Commit with feat: type
git add .
git commit -m "feat: add router enhancement"
git push -u origin feat/trigger-release
```

**Expected behavior:**
1. Auto-PR creates PR
2. CI runs tests
3. PR merges to main
4. Release workflow creates release PR "chore(main): release 0.2.0"
5. Release PR auto-merges
6. GitHub release v0.2.0 is created with binaries

## 📋 Phase 4: Verification Checklist

Use this checklist to verify everything works:

- [ ] PAT secret is created and valid
- [ ] Branch protection rules are configured
- [ ] Auto-merge is enabled in repository settings
- [ ] Required status check `test` is added to branch protection
- [ ] Test branch triggers auto-PR creation
- [ ] CI workflow runs on PRs
- [ ] CI skips when only docs change
- [ ] PRs auto-merge after CI passes
- [ ] Branches are auto-deleted after merge
- [ ] Release-please creates release PRs
- [ ] Release PRs auto-merge
- [ ] GitHub releases are created with binaries
- [ ] Cannot push directly to main (branch protection works)

## 🐛 Troubleshooting

### Issue: "Resource not accessible by integration"

**Cause:** Workflow needs PAT but is using GITHUB_TOKEN

**Solution:** Verify PAT secret exists:
```bash
gh secret list
```

### Issue: CI not running

**Cause:** Workflow file has syntax errors

**Solution:**
```bash
# Install actionlint
brew install actionlint

# Check workflow syntax
actionlint .github/workflows/*.yml
```

### Issue: Auto-merge not working

**Cause:** Branch protection not configured correctly

**Solution:**
1. Go to Settings → Branches
2. Verify required checks include `test`
3. Verify auto-merge is enabled
4. Verify github-actions bot can bypass PR requirement

### Issue: Release-please not creating PR

**Cause:** Commits don't follow conventional commits

**Solution:**
```bash
# Check commit format
git log -1 --pretty=%s

# Should match pattern: type: description
# Valid types: feat, fix, docs, test, refactor, chore
```

## 📚 Documentation

For detailed information about the automation system, see:

- **`internal/docs/WORKFLOWS.md`** - Complete workflow documentation
- **Conventional Commits:** https://www.conventionalcommits.org/
- **Release-Please:** https://github.com/googleapis/release-please
- **GitHub Actions:** https://docs.github.com/en/actions

## 🎯 Next Steps

After completing setup:

1. Delete this setup file (or commit it for future reference)
2. Create your first real feature branch
3. Let the automation handle PR creation and merging
4. Monitor the first few runs to ensure everything works
5. Celebrate your fully automated workflow! 🎉

## 🔄 Maintenance

### Monthly Tasks

- [ ] Check PAT expiration date
- [ ] Review workflow run history for issues
- [ ] Update GitHub Actions versions if needed

### Before PAT Expiration

1. Create new PAT with same permissions
2. Update PAT secret in repository
3. Verify workflows still work
4. Delete old PAT from GitHub settings

### Updating Workflows

To modify workflows:

1. Create feature branch
2. Edit workflow files
3. Push and create PR
4. Test changes on PR
5. Merge when verified

Remember: Workflow changes take effect immediately on push to main, so test thoroughly!
