# dnote-pg2sqlite

Migration tool from PostgreSQL to SQLite for Dnote.

## Installation

```bash
go install github.com/dnote/dnote-pg2sqlite@latest
```

## Usage

```bash
dnote-pg2sqlite \
  --pg-host localhost \
  --pg-port 5432 \
  --pg-database dnote \
  --pg-user dnote \
  --pg-password yourpassword \
  --sqlite-path /path/to/server.db
```

## Backup First

**Always backup PostgreSQL before migrating:**

```bash
pg_dump -h localhost -U dnote -d dnote > dnote_backup.sql
```

## What Gets Migrated

- Users & accounts
- Books & notes
- Sessions & tokens
- Notifications & email preferences

Full-text search index rebuilds automatically.

## Verification

After migration:

1. Check the tool's output for record counts
2. Start Dnote and verify login works
3. Verify notes and search work correctly
