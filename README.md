# dnote-pg2sqlite

Migration tool from PostgreSQL to SQLite for Dnote.

This tool migrates data from **Dnote server v2.x** (PostgreSQL) to **Dnote server v3** (SQLite).

## Prerequisites

- You must be running **Dnote server v2.x**

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
  --sqlite-path ~/.local/share/dnote/server.db
```

**Note**: Dnote v3 uses XDG directories. The default SQLite database path is `~/.local/share/dnote/server.db` (or `$XDG_DATA_HOME/dnote/server.db`).

**Safety**: The migration tool will refuse to run if the SQLite file already exists, preventing accidental overwrites. Remove the existing file if you need to re-run the migration.

## Backup First

**Always backup PostgreSQL before migrating:**

```bash
pg_dump -h localhost -U dnote -d dnote > dnote_backup.sql
```

## What Gets Migrated

- Users & accounts
- Books & notes
- Sessions & tokens

Full-text search index rebuilds automatically.

## Migration Workflow

1. **Ensure you're on v2.x**: Upgrade to Dnote server v2.x if needed
2. **Stop your Dnote server** to ensure data consistency during migration
3. **Backup your PostgreSQL database** (see above)
4. **Run the migration tool** with your PostgreSQL credentials
5. **Verify the migration** succeeded (check record counts)
6. **Upgrade to Dnote v3** and configure it to use the new SQLite database

## Verification

After migration:

1. Check the tool's output for record counts
2. Start Dnote v3 and verify login works
3. Verify notes and search work correctly
