# EnvQuack 🦆

Environment Variable Drift Detective - Keep your `.env` files in sync!

```
 ___            ___                 _    
| __|_ ___ ___ / _ \ _  _ __ _ __ _ _| |__ 
| _|| ' \ V / | (_) | || / _' / _' | / /
|___|_||_\_/   \__\_\\_,_\__,_\__,_|_\_\
                                        
Environment Variable Drift Detective 🦆
```

> ⚠️ **Alpha Release Notice (v0.1.0-alpha.1)**  
> EnvQuack is currently in **early alpha**. Expect rapid changes, incomplete features, and potential breaking changes.  
> Feedback, issues, and feature requests are highly appreciated as we shape the tool's roadmap.

## What is EnvQuack?

EnvQuack is a CLI tool that helps you keep your environment variables synchronized across different files. It compares your `.env` file against `.env.example` and detects:

- **Missing variables**: Present in example but absent in your `.env`
- **Extra variables**: Present in your `.env` but not documented in example
- **Docker Compose issues**: Variables required by services but missing in env files
- **Dockerfile problems**: ARG/ENV mismatches and unused build arguments

## Features

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

### Prebuilt Binaries (Recommended)

Download the latest prebuilt binaries from the [Releases page](https://github.com/DuckDHD/EnvQuack/releases):

- [Linux (amd64)](https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack-linux)
- [macOS (arm64)](https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack-macos)
- [Windows (amd64)](https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack.exe)

```bash
# Linux/macOS quick install
curl -L https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack-linux -o envquack
chmod +x envquack
./envquack --help
```

### Using Go Install (Recommended)

```bash
go install github.com/DuckDHD/EnvQuack/cmd/envquack@v0.1.0-alpha.1
```

### From Source

```bash
git clone https://github.com/DuckDHD/EnvQuack
cd EnvQuack
go build -o envquack cmd/envquack/main.go
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/DuckDHD/EnvQuack/releases/tag/v0.1.0-alpha.1)

## Usage

### Check for differences

```bash
envquack check
```

Example output when issues are found:
```
   __
<(X )___   QUACK!
 ( ._> /
  '---'

QUACK! 🦆 Environment issues detected:

🔴 Missing variables (present in .env.example but not in .env):
  - DB_HOST
  - API_KEY
  - SECRET_TOKEN

🟡 Extra variables (present in .env but not in .env.example):
  - DEBUG_MODE
  - TEMP_VAR

(Your gopher-duck is angry. Fix your .env!)
```

When everything is aligned:
```bash
$ envquack check
✅ All envs aligned.
(Your gopher-duck is calm and happy.)
```

### Sync missing variables

Automatically add missing variables to your `.env` file:

```bash
envquack sync
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

Example output:
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

### Comprehensive audit

Run a full environment audit across all your Docker files:

```bash
envquack audit
```

This will check:
- `.env` vs `.env.example` consistency
- Docker Compose environment requirements
- Dockerfile ARG and ENV usage
- Missing env_file references

Example output:
```
🔍 Running comprehensive environment audit...

📋 Checking .env vs .env.example:
  ✅ Basic env check passed

🐳 Checking docker-compose environment requirements:
  ✅ Docker Compose check passed

🐋 Checking Dockerfile environment requirements:
  🔴 Variables required by Dockerfile but missing in env files:
    - BUILD_VERSION
    - REDIS_URL

  🟠 ARG variables declared but never used:
    - UNUSED_BUILD_ARG

   __
<(X )___   QUACK!
 ( ._> /
  '---'
QUACK! 🦆 Audit found issues that need attention!
```

### Custom file paths

```bash
envquack check --env .env.local --example .env.template
envquack sync --env config/.env --example config/.env.example
envquack audit --compose docker-compose.prod.yml --dockerfile Dockerfile.prod
```

### Options

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

### Complete Docker setup workflow

1. Create your environment files:

```bash
# .env.example - Document all required variables
NODE_ENV=production
API_URL=https://api.example.com
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
SECRET_KEY=your_secret_key_here
REDIS_URL=redis://localhost:6379
```

2. Create your Dockerfile with proper ARG/ENV usage:

```dockerfile
# Build arguments
ARG NODE_ENV=production
ARG API_URL
ARG SECRET_KEY

# Runtime environment
ENV NODE_ENV=${NODE_ENV}
ENV API_BASE_URL=${API_URL}  
ENV JWT_SECRET=${SECRET_KEY}
```

3. Set up docker-compose.yml:

```yaml
services:
  web:
    build:
      args:
        - NODE_ENV=${NODE_ENV}
        - API_URL=${API_URL}
        - SECRET_KEY=${SECRET_KEY}
    environment:
      - DATABASE_URL=${DATABASE_URL}
    env_file: .env
```

4. Run comprehensive audit:

```bash
envquack audit --verbose
```

5. Sync missing variables:

```bash
envquack sync
```

6. Fill in actual values in your `.env` file.

### Basic workflow

1. Create your `.env.example` with all required variables:
```bash
# .env.example
DB_HOST=localhost
DB_PORT=5432
API_KEY=your_api_key_here
SECRET_TOKEN=your_secret_here
```

2. Check if your `.env` is complete:
```bash
envquack check
```

3. Sync missing variables:
```bash
envquack sync
```

4. Fill in the actual values in your `.env` file.

### CI/CD Integration

Add to your CI pipeline to ensure env files stay in sync:

```yaml
# GitHub Actions example
- name: Check env files
  run: |
    go install github.com/DuckDHD/EnvQuack/cmd/envquack@v0.1.0-alpha.1
    envquack check --no-duck
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

### Project Structure

```
envquack/
├── cmd/envquack/main.go      # CLI entrypoint
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
└── README.md
```

### Running tests

```bash
go test ./...
```

### Building

```bash
go build -o envquack cmd/envquack/main.go
```

## Roadmap

- ✅ **v0.1.0-alpha.1**: Basic .env comparison, sync, Docker Compose and Dockerfile support
- 🚧 **v0.1.0**: Stable release with bug fixes and polish
- 📋 **v0.2.0**: Kubernetes ConfigMap/Secret support
- 🎯 **v1.0.0**: Central schema files and multi-environment support

## Alpha Release Notes

This is **v0.1.0-alpha.1** - our first public release! 🎉

**What works well:**
- ✅ Basic .env comparison and sync
- ✅ Docker Compose environment analysis  
- ✅ Dockerfile ARG/ENV parsing
- ✅ Comprehensive audit across all sources
- ✅ Beautiful CLI output with duck art 🦆

**What might have rough edges:**
- ⚠️ Limited test coverage (we're working on it!)
- ⚠️ Some edge cases in complex configurations
- ⚠️ Error messages could be more helpful
- ⚠️ Performance not optimized for huge files

**Help us improve!**
- 🐛 [Report bugs](https://github.com/DuckDHD/EnvQuack/issues)
- 💡 [Request features](https://github.com/DuckDHD/EnvQuack/issues)
- 🤝 [Contribute code](https://github.com/DuckDHD/EnvQuack/pulls)
- ⭐ Star the repo if you find it useful!

Your feedback will directly shape the stable v0.1.0 release. Thank you for being an early adopter! 🦆

We welcome contributions! This is an alpha release, so there's lots of room for improvement.

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Submit a pull request

### Development Setup

```bash
git clone https://github.com/DuckDHD/EnvQuack
cd EnvQuack
go mod tidy
go build -o envquack cmd/envquack/main.go
```

### Areas That Need Help

- [ ] Unit tests for all parsers
- [ ] Integration tests
- [ ] Windows compatibility testing
- [ ] Performance optimization for large files
- [ ] Better error messages
- [ ] Documentation improvements

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

---

Made with 🦆 and ❤️ for developers who like their environment variables organized!