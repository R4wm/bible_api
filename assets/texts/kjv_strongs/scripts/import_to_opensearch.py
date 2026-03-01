#!/usr/bin/env python3
"""Import kjv_strongs.txt into OpenSearch.

TODO:
- Parse verse lines into canonical JSON docs
- Bulk index into target index with deterministic IDs
- Emit failed document report and summary
"""

from pathlib import Path
import argparse


def main() -> int:
    parser = argparse.ArgumentParser(description="Import kjv_strongs into OpenSearch")
    parser.add_argument("--url", default="http://localhost:9200", help="OpenSearch URL")
    parser.add_argument("--index", default="kjv_strongs_verses", help="Target index")
    parser.add_argument("--input", default=str(Path(__file__).resolve().parents[1] / "kjv_strongs.txt"), help="Input text file")
    args = parser.parse_args()

    print(f"[TODO] import {base} into OpenSearch")
    print(f"url={args.url} index={args.index} input={args.input}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
