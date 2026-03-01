#!/usr/bin/env python3
"""Verify checksum and basic import expectations for tyndale."""

from pathlib import Path
import argparse
import hashlib


def md5_file(path: Path) -> str:
    h = hashlib.md5()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def main() -> int:
    parser = argparse.ArgumentParser(description="Verify tyndale source checksum")
    parser.add_argument("--input", default=str(Path(__file__).resolve().parents[1] / "tyndale.txt"), help="Input text file")
    args = parser.parse_args()

    path = Path(args.input)
    print(f"md5({path.name})={md5_file(path)}")
    print("[TODO] add SQLite/OpenSearch row count verification")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
