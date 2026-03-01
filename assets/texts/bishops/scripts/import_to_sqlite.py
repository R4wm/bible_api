#!/usr/bin/env python3
"""Import bishops.txt into SQLite.

TODO:
- Parse verse lines into (book, chapter, verse, text)
- Insert into target table with idempotent upsert behavior
- Emit import summary for auditability
"""

from pathlib import Path
import argparse


def main() -> int:
    parser = argparse.ArgumentParser(description="Import bishops into SQLite")
    parser.add_argument("--db", required=True, help="Path to SQLite database")
    parser.add_argument("--table", default="bishops_verses", help="Target table")
    parser.add_argument("--input", default=str(Path(__file__).resolve().parents[1] / "bishops.txt"), help="Input text file")
    args = parser.parse_args()

    print(f"[TODO] import {base} into SQLite")
    print(f"db={args.db} table={args.table} input={args.input}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
