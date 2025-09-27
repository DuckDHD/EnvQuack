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
- 🎨 **Beautiful output**: ASCII duck art and colored reports

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
1. Read your `.env.example` file
2. Check what's missing in `.env`
3. Add missing variables with empty values
4. Preserve existing variables

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

- `--env`: Path to your env file (default: `.env`)
- `--example`: Path to your example file (default: `.env.example`)
- `--compose`: Path to docker-compose file (default: `docker-compose.yml`)
- `--dockerfile`: Path to Dockerfile (default: `Dockerfile`)
- `--verbose`, `-v`: Verbose output (shows additional details like unused ARGs)
- `--no-color`: Disable colored output
- `--no-duck`: Disable ASCII duck art (for serious environments)

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

## Development

### Project Structure

```
envquack/
├── cmd/envquack/main.go      # CLI entrypoint
├── internal/
│   ├── parser/
│   │   ├── env.go           # .env file parser
│   │   ├── compose.go       # docker-compose.yml parser  
│   │   └── dockerfile.go    # Dockerfile parser
│   ├── checker/
│   │   ├── diff.go          # Environment comparison logic
│   │   ├── compose.go       # Docker Compose analysis
│   │   ├── dockerfile.go    # Dockerfile analysis
│   │   └── report.go        # Report generation
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