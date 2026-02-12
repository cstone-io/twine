# CLI Versioning Strategy

## How It Works

The Twine CLI uses **build-time injection** to set version information. This is the standard approach used by production Go CLI tools like `kubectl`, `docker`, `gh`, etc.

### Variables in `main.go`

```go
var (
    // Build-time variables (set via -ldflags)
    version   = "dev"      // Default for development
    commit    = "none"     // Git commit hash
    date      = "unknown"  // Build timestamp
    builtBy   = "unknown"  // Who built it
)
```

These have **default values** for development but get **overridden at build time**.

## Building with Version Info

### Development Build (defaults)
```bash
go build ./cmd/twine
./twine version
# Output:
# Twine CLI
#   Version:    dev
#   Commit:     none
#   Built:      unknown
#   Built by:   unknown
```

### Production Build (with Makefile)
```bash
make build-cli
./dist/twine version
# Output:
# Twine CLI
#   Version:    v0.1.0
#   Commit:     a3f2c1b
#   Built:      2026-02-08_02:20:42
#   Built by:   cstone
```

### Manual Build with ldflags
```bash
go build -ldflags "\
  -X main.version=v1.0.0 \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.date=$(date -u '+%Y-%m-%d_%H:%M:%S') \
  -X main.builtBy=$(whoami)" \
  -o twine ./cmd/twine
```

## Makefile Targets

### `make build-cli`
Build for current platform with version info:
```bash
make build-cli
# → dist/twine (with version from git tag or "dev")
```

### `make install-cli`
Install to `$GOPATH/bin` with version info:
```bash
make install-cli
# → $GOPATH/bin/twine (usually ~/go/bin/twine)
```

### `make build-cli-all`
Build for all platforms:
```bash
make build-cli-all
# → dist/twine-darwin-amd64
# → dist/twine-darwin-arm64
# → dist/twine-linux-amd64
# → dist/twine-linux-arm64
# → dist/twine-windows-amd64.exe
```

### `make version-info`
Show what version info will be injected:
```bash
make version-info
# Version: v0.1.0
# Commit:  a3f2c1b
# Date:    2026-02-08_02:20:42
# Built by: cstone
```

## How Version is Determined

The Makefile automatically determines the version from **Git tags**:

```bash
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
```

**Examples:**
- No git repo: `dev`
- No tags: `a3f2c1b` (commit hash)
- Tagged: `v0.1.0`
- Tagged + 3 commits: `v0.1.0-3-ga3f2c1b`
- Uncommitted changes: `v0.1.0-dirty`

## Release Workflow

### 1. **Create a Git Tag**
```bash
git tag v0.1.0
git push origin v0.1.0
```

### 2. **GitHub Actions Automatically**
The `.github/workflows/release.yml` workflow:
- Builds binaries for all platforms
- Injects version info from the tag
- Creates a GitHub Release
- Attaches binaries to the release

### 3. **Users Download from GitHub Releases**
```bash
# Download for their platform
curl -L https://github.com/cstone/twine/releases/download/v0.1.0/twine-darwin-arm64 -o twine
chmod +x twine
./twine version
# Twine CLI
#   Version:    v0.1.0
#   Commit:     a3f2c1b
#   Built:      2026-02-08_02:20:42
#   Built by:   GitHub Actions
```

## Why This Approach?

### ✅ Advantages

1. **No hardcoded versions** - Version comes from git tags
2. **Traceable** - Commit hash tells you exactly what code was built
3. **Auditable** - Build date and builder for security/compliance
4. **Standard practice** - Same as kubectl, docker, gh, etc.
5. **CI/CD friendly** - Easy to automate in GitHub Actions

### ❌ Alternatives (and why they're worse)

**Hardcoded version in source:**
```go
const version = "0.1.0" // ❌ Must manually update
```
- Requires manual updates
- Easy to forget
- Source of truth is in code, not git tags

**Version file:**
```
VERSION.txt:
0.1.0
```
- Extra file to maintain
- Can get out of sync with git tags
- Still manual updates

**go.mod version:**
```go
module github.com/cstone/twine v0.1.0 // ❌ This is for module version, not app version
```
- `go.mod` version is for the **module API**, not the CLI tool
- They can differ (module v1.0.0 might contain CLI v0.5.3)

## Real-World Examples

### kubectl
```bash
$ kubectl version --client
Client Version: v1.29.1
Kustomize Version: v5.0.4-0.20230601165947-6ce0bf390ce3
```

### docker
```bash
$ docker version
Client:
 Version:           24.0.7
 API version:       1.43
 Go version:        go1.21.5
 Git commit:        afdd53b
 Built:             Thu Jan 11 11:22:05 2024
```

### gh (GitHub CLI)
```bash
$ gh version
gh version 2.42.1 (2024-01-09)
https://github.com/cli/cli/releases/tag/v2.42.1
```

All use build-time injection!

## Quick Reference

| Command | Purpose |
|---------|---------|
| `make build-cli` | Build with version info (local use) |
| `make install-cli` | Install to $GOPATH/bin |
| `make build-cli-all` | Build for all platforms (releases) |
| `make version-info` | See what will be injected |
| `git tag v1.0.0` | Create a version tag |
| `git push origin v1.0.0` | Trigger release workflow |

## Development Workflow

**Daily development:**
```bash
go install ./cmd/twine  # Quick install (version: "dev")
twine init myapp
```

**Building a release:**
```bash
git tag v0.2.0
git push origin v0.2.0
# GitHub Actions builds and creates release
```

**Local testing of release build:**
```bash
make build-cli
dist/twine version  # Check version info
```

## Summary

✅ **Do:** Use `-ldflags` to inject version at build time
✅ **Do:** Get version from `git describe --tags`
✅ **Do:** Include commit hash and build date
❌ **Don't:** Hardcode version in source
❌ **Don't:** Use `go.mod` version for CLI version
❌ **Don't:** Manually update version files
