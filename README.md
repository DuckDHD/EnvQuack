# EnvQuack 🦆

Environment Variable Drift Detective – Keep your `.env` files in sync!

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/DuckDHD/EnvQuack?include_prereleases&sort=semver)](https://github.com/DuckDHD/EnvQuack/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/DuckDHD/EnvQuack)](https://goreportcard.com/report/github.com/DuckDHD/EnvQuack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/DuckDHD/EnvQuack/issues)



```
 ___            ___                 _    
| __|_ ___ ___ / _ \ _  _ __ _ __ _ _| |__ 
| _|| ' \ V / | (_) | || / _' / _' | / /
|___|_||_\_/   \__\_\\_,_\__,_\__,_|_\_\
                                        
Environment Variable Drift Detective 🦆
```

> ⚠️ **Alpha Release Notice (v0.1.0-alpha.1)**  
> EnvQuack is in **early alpha**. Expect rapid changes, incomplete features, and breaking changes.  
> Feedback and bug reports are highly appreciated as we shape the stable v0.1.0 release.

---

## What is EnvQuack?

EnvQuack is a CLI tool that keeps your environment variables synchronized across files. It compares `.env` files against `.env.example` and detects:

- **Missing variables**: Present in example but missing in `.env`.
- **Extra variables**: Present in `.env` but not documented.
- **Docker Compose issues**: Variables required by services but missing in env files.
- **Dockerfile issues**: ARG/ENV mismatches and unused build arguments.

---

## Features

- 🦆 **Basic env checking**: Compare `.env` vs `.env.example`
- 🐳 **Docker Compose support**: Analyze `environment` and `env_file` usage
- 🐋 **Dockerfile analysis**: Parse ARG and ENV instructions
- 🔍 **Comprehensive audit**: Check all sources in one command
- 🔄 **Auto-sync**: Add missing variables to `.env` automatically
- 🎨 **Beautiful output**: ASCII duck art and color-coded reports

---

## Supported Platforms

| OS      | Architecture | Status |
|---------|--------------|--------|
| Linux   | amd64        | ✅ |
| macOS   | arm64        | ✅ |
| Windows | amd64        | ✅ (needs more testing) |

---

## Installation

### macOS (arm64)
```bash
curl -L https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack-macos -o envquack
chmod +x envquack && sudo mv envquack /usr/local/bin/
envquack --help
```

### Linux (amd64)
```bash
curl -L https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack-linux -o envquack
chmod +x envquack && sudo mv envquack /usr/local/bin/
envquack --help
```

### Windows (amd64, PowerShell)
```powershell
Invoke-WebRequest -Uri https://github.com/DuckDHD/EnvQuack/releases/download/v0.1.0-alpha.1/envquack.exe -OutFile envquack.exe
.\envquack.exe --help
```

### Verify checksum
```bash
# macOS/Linux
shasum -a 256 /usr/local/bin/envquack
# Compare with envquack-<os>.sha256 file from release
```

### Install with Go
```bash
go install github.com/DuckDHD/EnvQuack/cmd/envquack@v0.1.0-alpha.1
```

---

## Quickstart

### 1. Create `.env.example`
```bash
NODE_ENV=production
API_URL=https://api.example.com
DATABASE_URL=postgres://user:pass@localhost:5432/mydb
SECRET_KEY=your_secret_key
```

### 2. Run a check
```bash
envquack check
```

Example output with issues:
```
   __
<(X )___   QUACK!
 ( ._> /
  '---'

QUACK! 🦆 Environment issues detected:

🔴 Missing variables:
  - DB_HOST
  - API_KEY

🟡 Extra variables:
  - DEBUG_MODE
```

When everything is aligned:
```bash
✅ All envs aligned. (Your gopher-duck is calm and happy.)
```

---

## Commands

### `check`
Check for differences between `.env` and `.env.example`.
```bash
envquack check
```

### `sync`
Add missing variables to `.env` with empty values.
```bash
envquack sync
```

### `audit`
Run a full environment audit:
```bash
envquack audit
```

Checks:
- `.env` vs `.env.example` consistency
- Docker Compose env requirements
- Dockerfile ARG/ENV usage

---

## Options

| Option            | Default                | Description |
|-------------------|------------------------|-------------|
| `--env`           | `.env`                 | Path to your env file |
| `--example`       | `.env.example`         | Path to your example file |
| `--compose`       | `docker-compose.yml`   | Path to docker-compose file |
| `--dockerfile`    | `Dockerfile`           | Path to Dockerfile |
| `-v, --verbose`   | Off                     | Show unused ARGs and extra info |
| `--no-color`      | Off                     | Disable colored output |
| `--no-duck`       | Off                     | Disable ASCII duck art |

---

## Example Workflow

```bash
envquack check
envquack sync
envquack audit --verbose
```

---

## Project Structure

```
envquack/
├── cmd/envquack/main.go      # CLI entrypoint
├── internal/
│   ├── parser/               # File parsers
│   │   ├── env.go
│   │   ├── compose.go
│   │   └── dockerfile.go
│   ├── checker/              # Analysis logic
│   │   ├── diff.go
│   │   ├── compose.go
│   │   ├── dockerfile.go
│   │   └── report.go
│   ├── cli/commands.go       # CLI command bindings
│   └── quack/ascii.go        # ASCII art & messages
├── go.mod
└── README.md
```

---

## Roadmap

- ✅ **v0.1.0-alpha.1** – Initial alpha release  
- 🚧 **v0.1.0** – Stable release with bug fixes & polish  
- 📋 **v0.2.0** – Kubernetes ConfigMap/Secret support  
- 🎯 **v1.0.0** – Central schema files & multi-environment support  

---

## Alpha Release Notes

**What works well:**
- Basic `.env` comparison and sync
- Docker Compose and Dockerfile support
- Comprehensive audits
- Beautiful CLI with duck art 🦆

**Known Limitations:**
- Limited test coverage
- Some edge cases in complex configurations
- Error messages need improvement
- Performance not yet optimized

---

## Contributing

We welcome contributions!  

1. Fork the repository  
2. Create a feature branch: `git checkout -b feature/amazing-feature`  
3. Make your changes  
4. Add tests if possible  
5. Commit and push  
6. Submit a pull request  

Issues and feature requests are tracked in [GitHub Issues](https://github.com/DuckDHD/EnvQuack/issues).

---

## License

MIT License – see [LICENSE](LICENSE) for details.

---

Made with 🦆 and ❤️ for developers who like their environment variables tidy!
