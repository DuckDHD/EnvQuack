# EnvQuack 🦆

Environment Variable Drift Detective - Keep your `.env` files in sync!

```
 ___            ___                 _    
| __|_ ___ ___ / _ \ _  _ __ _ __ _ _| |__ 
| _|| ' \ V / | (_) | || / _' / _' | / /
|___|_||_\_/   \__\_\\_,_\__,_\__,_|_\_\
                                        
Environment Variable Drift Detective 🦆
```

## What is EnvQuack?

EnvQuack is a CLI tool that helps you keep your environment variables synchronized across different files. It compares your `.env` file against `.env.example` and detects:

- **Missing variables**: Present in example but absent in your `.env`
- **Extra variables**: Present in your `.env` but not documented in example

## Installation

### From Source

```bash
git clone https://github.com/yourusername/envquack
cd envquack
go build -o envquack cmd/envquack/main.go
```

### Using Go Install

```bash
go install github.com/yourusername/envquack/cmd/envquack@latest
```

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

### Custom file paths

```bash
envquack check --env .env.local --example .env.template
envquack sync --env config/.env --example config/.env.example
```

### Options

- `--env`: Path to your env file (default: `.env`)
- `--example`: Path to your example file (default: `.env.example`)
- `--verbose`, `-v`: Verbose output
- `--no-color`: Disable colored output
- `--no-duck`: Disable ASCII duck art (for serious environments)

## Examples

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
    go install github.com/yourusername/envquack/cmd/envquack@latest
    envquack check --no-duck
```

## Development

### Project Structure

```
envquack/
├── cmd/envquack/main.go      # CLI entrypoint
├── internal/
│   ├── parser/env.go         # .env file parser
│   ├── checker/
│   │   ├── diff.go          # Environment comparison logic
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

- ✅ **v0.1.0**: Basic .env comparison and sync
- 🚧 **v0.2.0**: Docker Compose and Dockerfile support
- 📋 **v0.3.0**: Kubernetes ConfigMap/Secret support
- 🎯 **v1.0.0**: Central schema files and multi-environment support

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

---

Made with 🦆 and ❤️ for developers who like their environment variables organized!