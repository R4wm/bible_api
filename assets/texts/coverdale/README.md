# coverdale

- Full name: Coverdale Bible
- Source file: `coverdale.txt`
- Upstream source: https://sourceforge.net/projects/biblesuper/files/All%20Bibles%20-%20Plain%20Text/EN-English/?utm_source=chatgpt.com
- Integrity file: `MD5SUM`

## Import Scripts

- `scripts/import_to_sqlite.py`: import this text into a SQLite database.
- `scripts/import_to_opensearch.py`: import this text into an OpenSearch index.
- `scripts/verify_import.py`: verify source checksum and basic record counts.

## Notes

This directory is intentionally self-contained so import logic and change history for this translation stay together.
