# Enhanced Error Messages Demo

This document demonstrates the enhanced error messages implemented in EnvQuack.

## Features

✅ **Line numbers** - Know exactly where the error is
✅ **Context lines** - See 2-3 lines before and after
✅ **Helpful hints** - Get suggestions on how to fix
✅ **Color support** - Highlighted output (respects `--no-color`)
✅ **Smart validation** - Catches common mistakes early

## Example Error Messages

### 1. Missing Equals Sign

**Problem:** Line is missing the `=` separator

```bash
Failed to parse .env

  Line 4: missing '=' separator

    2 | DATABASE_URL=postgres://localhost
    3 | API_KEY=secret123
    4 | MISSING_EQUALS_SIGN
    5 | PORT=3000
    6 |

  Hint: Each line should be in format KEY=VALUE
```

### 2. Invalid Variable Name

**Problem:** Variable name starts with a number or contains invalid characters

```bash
Failed to parse .env

  Line 3: invalid variable name '123_INVALID_KEY' (use A-Z, 0-9, _)

    1 | # Configuration
    2 | DATABASE_URL=postgres://localhost
    3 | 123_INVALID_KEY=value
    4 | PORT=3000
    5 |

  Hint: Each line should be in format KEY=VALUE
```

### 3. Empty Variable Name

**Problem:** Line starts with `=` without a key

```bash
Failed to parse .env

  Line 2: empty variable name before '='

    1 | DATABASE_URL=postgres://localhost
    2 | =value_without_key
    3 | PORT=3000
    4 |

  Hint: Each line should be in format KEY=VALUE
```

### 4. Unclosed Quote

**Problem:** Quote is opened but never closed

```bash
Failed to parse .env

  Line 2: unclosed double quote

    1 | DATABASE_URL=postgres://localhost
    2 | API_KEY="unclosed_quote
    3 | PORT=3000
    4 |

  Hint: Each line should be in format KEY=VALUE
```

### 5. YAML Syntax Error (docker-compose.yml)

**Problem:** Invalid YAML indentation or structure

```bash
Failed to parse docker-compose.yml

  Line 8: mapping values are not allowed here

    6 |   web:
    7 |     environment:
    8 |       - DATABASE_URL: ${DATABASE_URL}
    9 |     ports:
   10 |       - "3000:3000"

  Hint: Check YAML syntax - ensure proper indentation and structure
```

### 6. Dockerfile Instruction Typo

**Problem:** Common typo in Dockerfile instruction

```bash
Failed to parse Dockerfile

  Line 5: Invalid instruction 'ARGS'

    3 | FROM node:18-alpine
    4 |
    5 | ARGS NODE_ENV=production
    6 |
    7 | RUN npm install

  Hint: Did you mean 'ARG'? See https://docs.docker.com/engine/reference/builder/
```

## Implementation Details

### Error Package Structure

```
internal/errors/
├── parse_error.go       # ParseError type and formatting
└── parse_error_test.go  # Comprehensive tests
```

### Key Components

1. **ParseError Type**
   - Stores filename, line number, column, message, hint
   - Maintains context lines before and after the error
   - Supports both colored and non-colored output

2. **Context Extraction**
   - Automatically extracts 2-3 lines before and after the error
   - Handles edge cases (start/end of file)
   - Preserves original line content for display

3. **Formatting**
   - Clean, readable output with aligned line numbers
   - Optional column pointer (`^`) for precise error location
   - Gray context lines, highlighted error line
   - Color-coded sections (red errors, yellow line numbers, cyan hints)

4. **Common Error Constructors**
   - `MissingEqualsError()` - Missing `=` separator
   - `InvalidKeyError()` - Invalid variable name
   - `EmptyKeyError()` - Empty key before `=`
   - `UnclosedQuoteError()` - Unclosed quotes
   - `YAMLIndentationError()` - YAML formatting issues
   - `DockerfileInstructionError()` - Dockerfile syntax errors

### Parser Integration

All three parsers now track line numbers:

1. **.env Parser** ([internal/parser/env.go](internal/parser/env.go:16))
   - Reads entire file to track lines
   - Validates each line with detailed error messages
   - Checks for: missing `=`, invalid keys, unclosed quotes

2. **Docker Compose Parser** ([internal/parser/compose.go](internal/parser/compose.go:37))
   - Extracts line numbers from YAML errors
   - Provides context for YAML syntax issues
   - Handles complex nested structures

3. **Dockerfile Parser** ([internal/parser/docker.go](internal/parser/docker.go:28))
   - Tracks line numbers during parsing
   - Detects common instruction typos (ARGS → ARG, ENVS → ENV)
   - Handles multi-line instructions with backslashes

### CLI Integration

The CLI commands ([internal/cli/commands.go](internal/cli/commands.go:482)) now use the `displayError()` helper:

- Automatically detects `ParseError` types
- Respects `--no-color` flag
- Displays enhanced formatting to stderr
- Falls back to standard error display for non-parse errors

## Testing

Comprehensive test coverage in [internal/errors/parse_error_test.go](internal/errors/parse_error_test.go):

- ✅ Basic error formatting
- ✅ Context extraction (start, middle, end of file)
- ✅ Column pointers
- ✅ Hints and suggestions
- ✅ Line number alignment
- ✅ Color vs no-color output
- ✅ All common error types

Run tests:
```bash
go test ./internal/errors/... -v
```

## Before vs After

### Before (Generic Error)

```bash
$ envquack check
Error: failed to parse .env: invalid syntax
```

**User thinks:** "What's wrong? Where? Which line?"

### After (Enhanced Error)

```bash
$ envquack check
Failed to parse .env

  Line 15: missing '=' separator

  13 | API_KEY=sk_live_abc123
  14 | DATABASE_URL=postgres://localhost
  15 | BROKEN LINE WITHOUT EQUALS
       ^
  16 | PORT=3000
  17 |

  Hint: Each line should be in format KEY=VALUE
```

**User immediately knows:**
- ✅ Exact line number (15)
- ✅ What the error is (missing `=`)
- ✅ Context (surrounding lines)
- ✅ How to fix it (format hint)

## Success Criteria

✅ All parsers track line numbers
✅ Enhanced error messages with context
✅ Color support (with `--no-color` option)
✅ Helpful hints for common mistakes
✅ Comprehensive test coverage
✅ CLI integration complete
✅ Documentation updated

## Impact

Enhanced error messages significantly improve user experience by:

1. **Reducing debugging time** - Users can immediately locate the problem
2. **Providing actionable guidance** - Hints suggest how to fix issues
3. **Professional appearance** - Clean, well-formatted output
4. **Better error handling** - Catches syntax errors that were previously ignored
5. **Educational** - Users learn proper syntax through hints

---

**Made with 🦆 and ❤️ for developers who value helpful error messages!**
