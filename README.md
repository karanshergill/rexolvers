# rexolvers

🚀 More resolvers = More resolved subdomains = More potential vulnerabilities to exploit!

A lightweight DNS resolver aggregation tool written in Go that collects DNS resolver IP addresses from multiple public and trusted sources. Supports both file and database storage for integration with other security tools.

## Features

- **Multi-source aggregation**: Fetches resolvers from multiple sources
- **Dual storage**: Save to text files and/or SQLite database
- **Type classification**: Separates public and trusted resolvers
- **Deduplication**: Automatically removes duplicate IP addresses
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Lightweight**: Single binary with minimal dependencies

## Installation

### Build from source
```bash
git clone https://github.com/karanshergill/rexolvers
cd rexolvers
go build -o rexolvers .
```

## Usage

### Basic Operations

**Process public resolvers:**
```bash
./rexolvers --public
```

**Process trusted resolvers:**
```bash
./rexolvers --trusted
```

**Process both types:**
```bash
./rexolvers --all
```

### Database Operations

**Save to database:**
```bash
./rexolvers --trusted --db
```

**Save only to database (skip file):**
```bash
./rexolvers --public --db --file=false
```

**List resolvers from database:**
```bash
# List trusted resolvers
./rexolvers --list=trusted

# List public resolvers
./rexolvers --list=public

# List all resolvers
./rexolvers --list=all
```

**Show database statistics:**
```bash
./rexolvers --stats
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--public` | Process public resolvers | false |
| `--trusted` | Process trusted resolvers | false |
| `--all` | Process both public and trusted resolvers | false |
| `--db` | Save resolvers to SQLite database | false |
| `--file` | Save resolvers to text files | true |
| `--list` | List resolvers from database (public\|trusted\|all) | "" |
| `--stats` | Show database statistics | false |

## Output Files

- `public_resolvers.txt`: Contains public DNS resolver IPs
- `trusted_resolvers.txt`: Contains trusted DNS resolver IPs
- `resolvers.db`: SQLite database with all resolvers

## Database Schema

The SQLite database uses the following schema:

```sql
CREATE TABLE resolvers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT UNIQUE NOT NULL,
    resolver_type TEXT NOT NULL,
    source_url TEXT,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Configuration

The tool uses a YAML configuration file located at:
- **Linux/macOS**: `$HOME/.config/rexolvers/config.yaml`
- **Windows**: `%APPDATA%\rexolvers\config.yaml`

### Default Sources

**Public Sources:**
- proabiral/Fresh-Resolvers
- trickest/resolvers
- janmasarik/resolvers
- Phasip/resolvers

**Trusted Sources:**
- trickest/resolvers-trusted

## Integration Examples

### Use with massdns
```bash
# Get all resolvers from database
./rexolvers --list=all > resolvers.txt
massdns -r resolvers.txt -t A domains.txt
```

### Use with subfinder
```bash
# Export trusted resolvers
./rexolvers --list=trusted > trusted_resolvers.txt
subfinder -d example.com -r trusted_resolvers.txt
```

### Query database directly
```bash
# Using sqlite3 command line
sqlite3 resolvers.db "SELECT COUNT(*) FROM resolvers WHERE resolver_type='trusted';"
```

## Examples

```bash
# Initial setup - fetch and save everything
./rexolvers --all --db

# Daily update - refresh database only
./rexolvers --all --db --file=false

# Export for specific tool
./rexolvers --list=trusted > my_resolvers.txt

# Check what's in database
./rexolvers --stats
```

## License

This project is licensed under the MIT License.
