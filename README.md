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

### From Source

```bash
git clone https://github.com/DuckDHD/EnvQuack
cd EnvQuack
go build -o envquack cmd/envquack/main.go
```

### Using Go Install

```bash
go install github.com/DuckDHD/EnvQuack/cmd/envquack@latest
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
    go install github.com/yourusername/envquack/cmd/envquack@latest
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

- ✅ **v0.1.0**: Basic .env comparison and sync
- ✅ **v0.1.1**: Docker Compose and Dockerfile support  
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