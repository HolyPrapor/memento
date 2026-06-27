# Storage v2 Migration

Documentation for the storage layer migration from flat files to SQLite.

## Why SQLite

SQLite was chosen because it provides ACID transactions and fast indexed queries
without requiring a separate database server. The FTS5 extension enables
full-text search capabilities that were previously impossible with flat files.

## Migration Steps

The migration followed these steps:
1. Create new SQLite tables with the v2 schema
2. Write a data conversion script to transfer existing data
3. Run both systems in parallel for one week
4. Switch all reads to SQLite
5. Deprecate flat file storage

## Known Issues

During the migration, we discovered that large datasets cause the conversion to
time out. The workaround is to process data in batches of 10,000 records. This
will be addressed in a future storage optimization.
