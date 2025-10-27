# EnvQuack 🦆

Environment Variable Drift Detective – Keep your `.env` files in sync!

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/DuckDHD/EnvQuack?sort=semver)](https://github.com/DuckDHD/EnvQuack/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/DuckDHD/EnvQuack)](https://goreportcard.com/report/github.com/DuckDHD/EnvQuack)
[![Test Coverage](https://img.shields.io/badge/coverage-75.6%25-brightgreen.svg)](TEST_COVERAGE_REPORT.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/DuckDHD/EnvQuack/pulls)
[![Production Ready](https://img.shields.io/badge/status-production--ready-success.svg)](https://github.com/DuckDHD/EnvQuack/releases/tag/v0.1.0)

```
 ___            ___                 _    
| __|_ ___ ___ / _ \ _  _ __ _ __ _ _| |__ 
| _|| ' \ V / | (_) | || / _' / _' | / /
|___|_||_\_/   \__\_\\_,_\__,_\__,_|_\_\
                                        
Environment Variable Drift Detective 🦆
```

> ✨ **Production Release v0.1.0**  
> EnvQuack is now **production-ready** with comprehensive security, automatic backups, and 75.6% test coverage.  
> Safe for use in real-world projects and CI/CD pipelines!

---

## What is EnvQuack?

EnvQuack is a production-grade CLI tool that keeps your environment variables synchronized and secure. It detects drift, prevents data loss, and protects your secrets.

**Key Capabilities:**
- 🔍 **Drift Detection**: Compare `.env` vs `.env.example`
- 🐳 **Docker Integration**: Validate Compose and Dockerfile environments
- 📦 **Automatic Backups**: Never lose data during sync operations
- 🔒 **Secret Protection**: Auto-masks sensitive values in output
- 💬 **Smart Errors**: Line numbers with helpful hints
- 🔐 **Security**: Path validation prevents traversal attacks

---

## Features

### Core Features
- ✅ **Basic env checking** - Compare `.env` vs `.env.example`
- ✅ **Auto-sync** - Add missing variables automatically
- ✅ **Docker Compose support** - Analyze `environment` and `env_file` usage
- ✅ **Dockerfile analysis** - Parse ARG and ENV instructions
- ✅ **Comprehensive audit** - Check all sources in one command

### Safety & Security Features (NEW in v0.1.0)
- 🔒 **Path validation** - Prevents directory traversal attacks
- 📦 **Automatic backups** - Timestamped backups before modifications
- 🔄 **Restore command** - Recover from any backup
- 🎭 **Sensitive value masking** - Protects secrets in output (20+ patterns)
- 💬 **Enhanced error messages** - Line numbers with context and hints
- 🛡️ **File size limits** - Prevents memory exhaustion attacks

### Developer Experience
- 🎨 **Beautiful output** - ASCII duck art and color-coded reports
- 📍 **Precise errors** - Know exactly where and what went wrong
- 🚀 **CI/CD ready** - Proper exit codes for automation
- 🎯 **Helpful hints** - Actionable suggestions for fixing issues

---

## Installation

### Quick Install

#### macOS (Intel & Apple Silicon)
```bash
# Intel
curl -L https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0/envquack-darwin-amd64 -o envquack

# Apple Silicon (M1/M2/M3)
curl -L https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0/envquack-darwin-arm64 -o envquack

chmod +x envquack && sudo mv envquack /usr/local/bin/
envquack --version
```

#### Linux (amd64)
```bash
curl -L https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0/envquack-linux-amd64 -o envquack
chmod +x envquack && sudo mv envquack /usr/local/bin/
envquack --version
```

#### Windows (amd64, PowerShell)
```powershell
Invoke-WebRequest -Uri https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0/envquack-windows-amd64.exe -OutFile envquack.exe
.\envquack.exe --version
```

### Install with Go
```bash
go install github.com/DuckDHD/EnvQuack/cmd/envquack@v0.1.0
```

### Verify Installation
```bash
envquack --version
# EnvQuack v0.1.0
```

---

## Quickstart

### 1. Create `.env.example`
```bash
NODE_ENV=production
API_URL=https://api.example.com
DATABASE_URL=postgres://user:pass@localhost:5432/mydb
API_KEY=your_api_key
```

### 2. Run a check
```bash
envquack check
```

**Example output with issues:**
```
   __
<(X )___   QUACK!
 ( ._> /
  '---'

QUACK! 🦆 Environment issues detected:

❌ Missing variables in .env:
  - API_KEY
  - DATABASE_URL

⚠️  Extra variables in .env:
  - DEBUG_MODE
  - OLD_CONFIG
```

**When everything is aligned:**
```bash
   __
<(o )___   
 ( ._> /
  '---'

✅ All envs aligned. The duck approves! 🦆
```

---

## Commands

### `check`
Check for differences between `.env` and `.env.example`.

```bash
envquack check
envquack check --env .env.local --example .env.template
```

**Exit codes:**
- `0` - No issues found
- `1` - Issues detected or error occurred

---

### `sync`
Add missing variables to `.env` with empty values. **Automatically creates a backup first!**

```bash
envquack sync
```

**Output:**
```
📦 Backup created: .env.backup.20241027-143052
   __
<(~ )___   Syncing...
 ( ._> /
  '---'

Adding 2 missing variables to .env:
  + API_KEY
  + DATABASE_URL

✅ Successfully synced 2 variables!
```

**Options:**
- `--no-backup` - Skip backup creation (not recommended)
- `--cleanup-backups` - Remove old backups (keeps 5 most recent)

---

### `restore`
Restore `.env` from a backup.

```bash
# List available backups
envquack restore --list

# Restore from most recent backup
envquack restore --latest

# Restore from specific backup
envquack restore .env.backup.20241027-143052
```

**Example output:**
```
Available backups for .env:
  1. .env.backup.20241027-143052 (2024-10-27 14:30:52, 234 bytes)
  2. .env.backup.20241027-120015 (2024-10-27 12:00:15, 198 bytes)
  3. .env.backup.20241026-093045 (2024-10-26 09:30:45, 187 bytes)
```

---

### `audit`
Run a comprehensive environment audit across all sources.

```bash
envquack audit
envquack audit --verbose
```

**Checks performed:**
1. `.env` vs `.env.example` consistency
2. Docker Compose environment requirements
3. Dockerfile ARG/ENV usage
4. Unused variables and hardcoded values (verbose mode)

---

## Advanced Usage

### Custom File Paths
```bash
envquack check --env config/.env --example config/.env.example
envquack audit --compose docker/compose.yml --dockerfile docker/Dockerfile
```

### CI/CD Integration
```bash
# In your CI pipeline
envquack check --no-duck --no-color
if [ $? -eq 1 ]; then
  echo "❌ Environment variables out of sync!"
  exit 1
fi
```

### Debug Mode
```bash
# Show actual sensitive values (WARNING: insecure)
envquack check --show-secrets --verbose
⚠️  WARNING: Showing actual sensitive values (insecure)
```

---

## Security Features

### Path Validation
EnvQuack prevents security vulnerabilities by validating all file paths:

**Blocked:**
- Directory traversal: `../../../etc/passwd`
- Absolute paths: `/etc/shadow`, `C:\Windows\System32`
- Paths outside working directory

**Allowed:**
- Relative paths in current directory: `.env`, `config/.env`
- Subdirectories: `docker/.env`, `app/config/.env`

**Override (advanced users only):**
```bash
envquack check --allow-unsafe-paths --env /opt/app/.env
⚠️  WARNING: Path validation disabled
```

---

### Sensitive Value Masking

EnvQuack automatically masks sensitive values in all output to prevent accidental exposure.

**What gets masked:**

**By Variable Name:**
- PASSWORD, PASSWD, SECRET, KEY, TOKEN
- API_KEY, APIKEY, PRIVATE, CREDENTIAL
- AUTH, CERTIFICATE, OAUTH, JWT, SESSION

**By Value Prefix:**
- Stripe: `sk_`, `pk_`, `rk_`
- GitHub: `ghp_`, `gho_`, `ghs_`, `ghr_`
- AWS: `AKIA`, `ASIA`
- Slack: `xoxb-`, `xoxp-`, `xoxa-`
- Google: `AIza`, `ya29.`

**By Pattern:**
- JWT tokens (3-part base64)
- Long base64 strings (32+ chars)
- Long hex strings (32+ chars)
- UUIDs

**Example:**
```bash
# Before masking (INSECURE):
Missing: API_KEY (example: sk_live_a1b2c3d4e5f6g7h8i9j0)

# After masking (SECURE):
Missing: API_KEY (example: sk********************j0)
```

**Format:** Shows first 2 and last 2 characters for recognition.

---

### Enhanced Error Messages

Get helpful, actionable error messages with line numbers and context:

```bash
# Instead of:
Error: failed to parse .env: invalid syntax

# You get:
Failed to parse .env

  Line 15: Missing '=' separator
  
  13 | API_KEY=sk_live_abc123
  14 | DATABASE_URL=postgres://localhost
  15 | BROKEN_LINE_WITHOUT_EQUALS
       ^
  16 | PORT=3000
  17 | DEBUG=true
  
  Hint: Each line should be in format KEY=VALUE
```

**Features:**
- Exact line numbers
- 2-3 lines of context
- Column pointers (^)
- Helpful hints
- Color-coded output

---

## Backup and Recovery

### Automatic Backups

Every `sync` operation automatically creates a timestamped backup:

```bash
$ envquack sync
📦 Backup created: .env.backup.20241027-143052
✓ Synced 3 variables to .env
```

**Backup format:** `.env.backup.YYYYMMDD-HHMMSS`

**Location:** Same directory as the original file

---

### Restore from Backup

**List available backups:**
```bash
envquack restore --list
```

**Restore from latest:**
```bash
envquack restore --latest
✅ Successfully restored .env from .env.backup.20241027-143052
```

**Restore from specific backup:**
```bash
envquack restore .env.backup.20241027-120015
```

---

### Backup Management

**Skip backup (not recommended):**
```bash
envquack sync --no-backup
```

**Auto-cleanup old backups:**
```bash
envquack sync --cleanup-backups
# Keeps 5 most recent, deletes backups older than 30 days
```

**Manual cleanup:**
```bash
# Remove backups older than 7 days
find . -name ".env.backup.*" -mtime +7 -delete
```

---

## Options

| Option                | Default                | Description |
|-----------------------|------------------------|-------------|
| `--env`               | `.env`                 | Path to your env file |
| `--example`           | `.env.example`         | Path to your example file |
| `--compose`           | `docker-compose.yml`   | Path to docker-compose file |
| `--dockerfile`        | `Dockerfile`           | Path to Dockerfile |
| `-v, --verbose`       | Off                    | Show unused ARGs and extra info |
| `--no-color`          | Off                    | Disable colored output |
| `--no-duck`           | Off                    | Disable ASCII duck art |
| `--no-backup`         | Off                    | Skip backup creation (dangerous) |
| `--cleanup-backups`   | Off                    | Auto-cleanup old backups |
| `--show-secrets`      | Off                    | Show actual sensitive values (insecure) |
| `--allow-unsafe-paths`| Off                    | Allow files outside working directory |

---

## Example Workflows

### Basic Development Workflow
```bash
# Check for drift
envquack check

# Sync missing variables
envquack sync

# Run comprehensive audit
envquack audit --verbose
```

### CI/CD Pipeline
```bash
#!/bin/bash
# .github/workflows/env-check.yml or similar

# Check environment variables
envquack check --no-duck --no-color
if [ $? -ne 0 ]; then
  echo "❌ Environment variables are out of sync!"
  echo "Run 'envquack sync' locally to fix."
  exit 1
fi

echo "✅ Environment variables are synchronized"
```

### Docker Project Audit
```bash
# Check everything in one command
envquack audit \
  --compose docker-compose.yml \
  --dockerfile Dockerfile \
  --verbose
```

### Recovery Workflow
```bash
# Oops, sync broke something!
envquack restore --list
envquack restore --latest

# Verify restoration
envquack check
```

---

## Project Structure

```
envquack/
├── cmd/envquack/main.go          # CLI entry point
├── internal/
│   ├── parser/                   # File parsers
│   │   ├── env.go                # .env parser
│   │   ├── compose.go            # docker-compose.yml parser
│   │   └── docker.go             # Dockerfile parser
│   ├── checker/                  # Analysis logic
│   │   ├── diff.go               # Basic comparison
│   │   ├── compose.go            # Compose validation
│   │   ├── docker.go             # Dockerfile validation
│   │   └── report.go             # Report generation
│   ├── cli/commands.go           # CLI command bindings
│   ├── security/                 # Security features (NEW)
│   │   └── validate.go           # Path validation
│   ├── backup/                   # Backup system (NEW)
│   │   └── backup.go             # Backup & restore
│   ├── masking/                  # Value masking (NEW)
│   │   └── mask.go               # Sensitive value detection
│   ├── errors/                   # Enhanced errors (NEW)
│   │   └── parse_error.go        # Error formatting
│   └── quack/ascii.go            # ASCII art & messages
├── go.mod
├── README.md
├── CHANGELOG.md                  # Release notes
└── TEST_COVERAGE_REPORT.md      # Coverage details
```

---

## Supported Platforms

| OS      | Architecture | Status             |
|---------|--------------|-------------------|
| Linux   | amd64        | ✅ Fully supported |
| macOS   | amd64 (Intel)| ✅ Fully supported |
| macOS   | arm64 (M1+)  | ✅ Fully supported |
| Windows | amd64        | ✅ Fully supported |

---

## Roadmap

### ✅ Completed
- **v0.1.0-alpha.1** - Initial alpha release
- **v0.1.0** - Production release with security, backups, testing

### 🚧 Planned

#### v0.1.1 (Quick Wins)
- Configuration file support (`.envquack.yml`)
- JSON/YAML output for CI/CD
- Duplicate variable detection
- Variable naming convention validation

#### v0.2.0 (Major Features)
- Multi-environment support (`.env.dev`, `.env.prod`, etc.)
- Type validation (PORT must be integer, URL format, etc.)
- Variable expansion (`${BASE_URL}/api`)
- Git integration (pre-commit hooks)
- Kubernetes ConfigMap/Secret support

#### v1.0.0 (Enterprise)
- Central schema files
- Team collaboration features
- API access
- Advanced reporting

See [ROADMAP.md](ROADMAP.md) for detailed feature plans.

---

## Testing & Quality

EnvQuack has comprehensive test coverage:

| Package | Coverage | Status |
|---------|----------|--------|
| Parser | 100% | ✅ Perfect |
| Masking | 100% | ✅ Perfect |
| Errors | 100% | ✅ Perfect |
| Quack | 100% | ✅ Perfect |
| Security | 87.9% | ✅ Excellent |
| Checker | 83.2% | ✅ Excellent |
| Backup | 75.0% | ✅ Good |
| **Total** | **75.6%** | ✅ **Production Ready** |

**Test Statistics:**
- 110+ test scenarios
- 13 test files
- All critical paths covered
- Cross-platform verified

See [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) for details.

---

## Contributing

We welcome contributions! 🎉

### How to Contribute

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes
4. **Add tests** (we maintain 75%+ coverage)
5. **Commit** with clear messages
6. **Push** to your fork
7. **Submit** a pull request

### Guidelines

- Write tests for new features
- Update documentation
- Follow existing code style
- Add entries to CHANGELOG.md
- Ensure all tests pass: `go test ./...`

### Areas We Need Help

- 🐛 Bug reports and fixes
- 📖 Documentation improvements
- ✨ Feature suggestions
- 🌐 Internationalization (i18n)
- 🎨 UI/UX improvements
- 🧪 More test coverage

See [open issues](https://github.com/DuckDHD/EnvQuack/issues) or [create a new one](https://github.com/DuckDHD/EnvQuack/issues/new).

---

## FAQ

### Q: Is EnvQuack safe for production use?
**A:** Yes! v0.1.0 is production-ready with 75.6% test coverage, automatic backups, and comprehensive security features.

### Q: Will EnvQuack modify my files without permission?
**A:** Only the `sync` command modifies files, and it always creates a backup first (unless you use `--no-backup`).

### Q: How does EnvQuack handle secrets?
**A:** EnvQuack automatically masks 20+ patterns of sensitive values (API keys, passwords, tokens) in all output.

### Q: Can I use EnvQuack in CI/CD?
**A:** Absolutely! EnvQuack has proper exit codes and supports `--no-color` and `--no-duck` for clean CI logs.

### Q: What if I need to access files outside my project directory?
**A:** Use `--allow-unsafe-paths` flag, but be cautious as this bypasses security validation.

### Q: Does EnvQuack send any data externally?
**A:** No. EnvQuack runs entirely locally and never sends data anywhere.

---

## Security

### Reporting Security Issues

If you discover a security vulnerability, please email **[your-email]** instead of using the issue tracker.

### Security Features

- ✅ Path traversal prevention
- ✅ File size limits (10MB default)
- ✅ Sensitive value masking
- ✅ No network access
- ✅ No command execution
- ✅ Input validation

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

Special thanks to:
- All alpha testers who provided feedback
- Contributors who reported issues and submitted PRs
- The Go community for excellent tooling and libraries

---

## Links

- **GitHub Repository**: https://github.com/DuckDHD/EnvQuack
- **Issue Tracker**: https://github.com/DuckDHD/EnvQuack/issues
- **Releases**: https://github.com/DuckDHD/EnvQuack/releases
- **Documentation**: https://github.com/DuckDHD/EnvQuack#readme
- **CHANGELOG**: [CHANGELOG.md](CHANGELOG.md)

---

<div align="center">

**Made with 🦆 and ❤️ for developers who like their environment variables tidy!**

[⬆ Back to Top](#envquack-)

</div>