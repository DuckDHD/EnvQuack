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

- 🦆 **Basic env checking**: Compare `.env` vs `.env.example`
- 🐳 **Docker Compose support**: Analyze `environment` and `env_file` usage
- 🐋 **Dockerfile analysis**: Parse ARG and ENV instructions
- 🔍 **Comprehensive audit**: Check all sources in one command
- 🔄 **Auto-sync**: Add missing variables to your `.env` automatically
- 📦 **Automatic backups**: Never lose data - backups created before every modification
- 🔙 **Easy restore**: Restore from timestamped backups with a single command
- 🔐 **Sensitive value masking**: Automatically masks secrets in terminal output and logs
- 🎨 **Beautiful output**: ASCII duck art and colored reports
- ⚡ **Helpful error messages**: Line numbers and context to quickly fix issues

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
git clone https://github.com/DuckDHD/EnvQuack
cd EnvQuack
go build -o envquack cmd/envquack/main.go
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

This will:
1. **Create a timestamped backup** of your `.env` file (automatic safety!)
2. Read your `.env.example` file
3. Check what's missing in `.env`
4. Add missing variables with empty values
5. Preserve existing variables

Example output:
```bash
$ envquack sync
📦 Backup created: .env.backup.20241027-143052
   __
<(~ )___   Syncing...
 ( ._> /
  '---'
Adding 2 missing variables to .env:
  + NEW_VARIABLE
  + ANOTHER_VAR

✅ Successfully synced 2 variables!
Don't forget to set the actual values in your .env file.
```

**Skip backup (not recommended):**
```bash
envquack sync --no-backup
```

**Automatically cleanup old backups:**
```bash
envquack sync --cleanup-backups
```
This keeps the 5 most recent backups and deletes backups older than 30 days.

### Backup and Recovery

EnvQuack automatically creates timestamped backups before modifying your `.env` file to prevent data loss.

#### List available backups

```bash
envquack restore --list
```

✅ Successfully synced 2 variables!
```
Available backups for .env:
  1. .env.backup.20241027-143052 (2024-10-27 14:30:52, 112 bytes)
  2. .env.backup.20241027-120000 (2024-10-27 12:00:00, 98 bytes)
  3. .env.backup.20241026-180000 (2024-10-26 18:00:00, 87 bytes)
```

#### Restore from most recent backup

```bash
envquack restore --latest
```

#### Restore from specific backup

```bash
envquack restore .env.backup.20241027-143052
```

#### Backup file format

Backups are stored in the same directory as your `.env` file with the format:
```
.env.backup.YYYYMMDD-HHMMSS
```

For example:
- `.env.backup.20241027-143052` = backup created on October 27, 2024 at 14:30:52

#### Safety features

- **Automatic backups**: Created before every `sync` operation (unless `--no-backup` is used)
- **Atomic operations**: Uses temporary files to prevent corruption
- **Permission preservation**: Backup files maintain the same permissions as the original
- **No data loss**: Even if sync fails, your original data is safe in the backup

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
envquack check --env .env.local --example .env.template
envquack sync --env config/.env --example config/.env.example
envquack audit --compose docker-compose.prod.yml --dockerfile Dockerfile.prod
```

---

#### Global Flags

- `--env`: Path to your env file (default: `.env`)
- `--example`: Path to your example file (default: `.env.example`)
- `--compose`: Path to docker-compose file (default: `docker-compose.yml`)
- `--dockerfile`: Path to Dockerfile (default: `Dockerfile`)
- `--verbose`, `-v`: Verbose output (shows additional details including masked example/actual values)
- `--no-color`: Disable colored output
- `--no-duck`: Disable ASCII duck art (for serious environments)
- `--allow-unsafe-paths`: Allow access to files outside current directory (USE WITH CAUTION)
- `--show-secrets`: Show actual sensitive values (WARNING: insecure, use only for debugging)

#### Sync Command Flags

- `--no-backup`: Skip creating backup before sync (not recommended)
- `--cleanup-backups`: Automatically cleanup old backups (keeps last 5, deletes >30 days)

#### Restore Command Flags

- `--list`: List all available backups
- `--latest`: Restore from most recent backup

## Examples

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

## Error Messages

EnvQuack provides clear, actionable error messages with line numbers and context to help you quickly identify and fix configuration issues.

### Example: Missing equals sign

When you have a syntax error in your `.env` file, EnvQuack shows you exactly where the problem is:

```bash
$ envquack check
Failed to parse .env

  Line 23: missing '=' separator

  21 | DATABASE_URL=postgres://localhost
  22 | API_KEY=secret123
  23 | BROKEN_LINE_WITHOUT_EQUALS
  24 | PORT=3000

  Hint: Each line should be in format KEY=VALUE
```

### Example: Invalid variable name

```bash
Failed to parse .env

  Line 5: invalid variable name '123_INVALID' (use A-Z, 0-9, _)

   3 | DATABASE_URL=postgres://localhost
   4 | API_KEY=secret123
   5 | 123_INVALID=value
   6 | PORT=3000

  Hint: Each line should be in format KEY=VALUE
```

### Example: Unclosed quote

```bash
Failed to parse .env

  Line 10: unclosed double quote

   8 | DATABASE_URL=postgres://localhost
   9 | API_KEY="secret123"
  10 | BROKEN="unclosed_quote
  11 | PORT=3000

  Hint: Each line should be in format KEY=VALUE
```

### Features

- **Line numbers**: Know exactly which line has the error
- **Context**: See 2-3 lines before and after the error
- **Helpful hints**: Get suggestions on how to fix common mistakes
- **Color support**: Errors are highlighted in color (unless `--no-color` is used)
- **Works for all file types**: `.env`, `docker-compose.yml`, and `Dockerfile`

## Security

EnvQuack implements multiple security features to protect your data and prevent vulnerabilities.

### Data Safety

- **Automatic Backups**: Creates timestamped backup before any modification
- **Atomic Operations**: Uses temporary files to prevent corruption during writes
- **Permission Preservation**: Backups maintain the same permissions as the original
- **No Data Loss**: Even if operations fail, your original data is safe

### Sensitive Value Masking

EnvQuack automatically masks sensitive values in all output to prevent accidental secret exposure in terminal output, CI/CD logs, screenshots, and screen shares.

#### What gets masked?

**Variable names containing:**
- `PASSWORD`, `SECRET`, `KEY`, `TOKEN`, `PRIVATE`, `CREDENTIAL`, `AUTH`
- `CERTIFICATE`, `CERT`, `JWT`, `SESSION`, `SALT`, `HASH`, `ENCRYPTION`

**Values with sensitive prefixes:**
- Stripe: `sk_`, `pk_`, `rk_` (secret, public, restricted keys)
- GitHub: `ghp_`, `gho_`, `ghs_`, `ghr_` (PAT, OAuth, server, refresh tokens)
- AWS: `AKIA`, `ASIA` (access keys, session tokens)
- Slack: `xoxb-`, `xoxp-`, `xoxa-` (bot, user, app tokens)
- Google: `AIza`, `ya29.` (API keys, OAuth)
- Generic: `key-`, `token-`, `Bearer `

**Value patterns:**
- JWT tokens (3-part base64)
- Long base64 strings (32+ chars)
- Long hex strings (32+ chars)
- UUIDs (might be used as secrets)

#### Example output with masking

```bash
# Verbose mode shows masked values for security
$ envquack check --verbose
Missing variables in .env:
  - API_KEY (example: sk********************89)
  - DATABASE_PASSWORD (example: My***************23)
  - GITHUB_TOKEN (example: gh*********************l2)

Extra variables in .env:
  - OLD_API_KEY (value: sk********************45)
```

**Format**: Shows first 2 and last 2 characters for recognition (e.g., `sk***89`)

#### Debug mode (show actual values)

For debugging purposes only, you can disable masking:

```bash
envquack check --show-secrets --verbose
⚠️  WARNING: Showing actual sensitive values (insecure)
Missing variables in .env:
  - API_KEY (example: sk_live_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9)
```

⚠️ **NEVER use `--show-secrets` in:**
- CI/CD pipelines
- Shared screens or screen recordings
- Public logs
- Production environments
- Any scenario where secrets might be captured

### Path Validation

By default, EnvQuack **only allows access to files within the current working directory**. This prevents:
- **Path traversal attacks** (accessing files outside the project directory)
- **Reading/modifying sensitive system files** (like `/etc/passwd`, `~/.ssh/id_rsa`)
- **Potential data exfiltration**

### Blocked Paths

The following types of paths are blocked by default:

```bash
# ❌ Parent directory traversal - BLOCKED
envquack check --env ../../../etc/passwd

# ❌ Absolute paths - BLOCKED
envquack check --env /etc/shadow
envquack sync --env C:\Windows\System32\config

# ❌ Hidden parent traversal - BLOCKED
envquack check --env config/../../secrets/api.key
```

### Allowed Paths

Only relative paths within the current directory are allowed:

```bash
# ✅ Current directory - ALLOWED
envquack check --env .env

# ✅ Subdirectories - ALLOWED
envquack check --env config/.env.local
envquack sync --env deployment/environments/.env.prod

# ✅ Nested subdirectories - ALLOWED
envquack audit --compose infra/docker/docker-compose.yml
```

### File Size Limits

To prevent Out-of-Memory (OOM) attacks, EnvQuack enforces a **10MB file size limit** by default. Files larger than this will be rejected with a clear error message.

### Bypassing Validation (Advanced)

In some cases, you may need to access files outside the current directory (e.g., system configurations, shared configs in `/etc/`).

You can use the `--allow-unsafe-paths` flag to bypass path validation:

```bash
# ⚠️  WARNING: Only use if you trust the file paths
envquack check --allow-unsafe-paths --env /opt/configs/.env
```

**When using this flag:**
- A warning message will be displayed
- You are responsible for ensuring the paths are safe
- This should only be used in trusted environments
- Never use this with user-provided input

### Best Practices

1. **Keep configs in your project directory**: Structure your project to keep all configuration files within the repository
2. **Use relative paths**: Always use relative paths like `./config/.env` instead of absolute paths
3. **Review CI/CD scripts**: Ensure your automation scripts don't use `--allow-unsafe-paths` unnecessarily
4. **Validate before deployment**: Run `envquack audit` before deploying to catch configuration issues early

## Development

## Project Structure

```
envquack/
├── cmd/envquack/main.go          # CLI entry point
├── internal/
│   ├── parser/
│   │   ├── env.go           # .env file parser
│   │   ├── compose.go       # docker-compose.yml parser
│   │   └── docker.go        # Dockerfile parser
│   ├── checker/
│   │   ├── diff.go          # Environment comparison logic
│   │   ├── compose.go       # Docker Compose analysis
│   │   ├── docker.go        # Dockerfile analysis
│   │   └── report.go        # Report generation
│   ├── backup/
│   │   └── backup.go        # Automatic backup and restore functionality
│   ├── masking/
│   │   └── mask.go          # Sensitive value masking for security
│   ├── security/
│   │   └── validate.go      # Path validation and security checks
│   ├── errors/
│   │   └── parse_error.go   # Enhanced error messages with context
│   ├── cli/commands.go      # Cobra CLI commands
│   └── quack/ascii.go       # ASCII art and messages
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