# Rexolvers

A lightweight DNS resolver aggregation tool that collects DNS resolver IP addresses from multiple public and trusted sources.

## Installation

### Build from source
```bash
git clone https://github.com/karanshergill/rexolvers
cd rexolvers
go build -o rexolvers .
```

### Docker
```bash
# Build the Docker image
docker build -t rexolvers .

# Or pull from a registry (if available)
docker pull your-registry/rexolvers:latest
```

## Usage

### Local Binary

#### Basic Operations

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

#### Database Operations

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

### Docker

Docker usage allows you to run the application in a containerized environment and export the database to your host machine.

#### Setup
```bash
# Create directory for database export
mkdir -p ./docker-output

# Build the Docker image
docker build -t rexolvers .
```

#### Basic Operations

**Process all resolvers and save to database:**
```bash
docker run -v $(pwd)/docker-output:/app/data rexolvers --all --db
```

**Process public resolvers only:**
```bash
docker run -v $(pwd)/docker-output:/app/data rexolvers --public --db
```

**Process trusted resolvers only:**
```bash
docker run -v $(pwd)/docker-output:/app/data rexolvers --trusted --db
```

#### Database Operations

**List resolvers from exported database:**
```bash
# List all resolvers
docker run -v $(pwd)/docker-output:/app/data rexolvers --list=all

# List trusted resolvers
docker run -v $(pwd)/docker-output:/app/data rexolvers --list=trusted

# List public resolvers
docker run -v $(pwd)/docker-output:/app/data rexolvers --list=public
```

**Show database statistics:**
```bash
docker run -v $(pwd)/docker-output:/app/data rexolvers --stats
```

#### Accessing Exported Database

After running with the `--db` flag, the SQLite database will be available at:
```
./docker-output/resolvers.db
```

You can then:
- Copy this database file anywhere you need it
- Query it directly with sqlite3: `sqlite3 ./docker-output/resolvers.db`
- Use it with other tools that accept SQLite databases

#### Docker Examples
```bash
# Complete workflow - fetch data and export database
docker build -t rexolvers .
mkdir -p ./docker-output
docker run -v $(pwd)/docker-output:/app/data rexolvers --all --db

# Check what was collected
docker run -v $(pwd)/docker-output:/app/data rexolvers --stats

# Export trusted resolvers for use with other tools
docker run -v $(pwd)/docker-output:/app/data rexolvers --list=trusted > trusted_resolvers.txt
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

## Integration Examples

### Use with massdns
```bash
# Get all resolvers from database
./rexolvers --list=all > resolvers.txt
massdns -r resolvers.txt -t A domains.txt
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