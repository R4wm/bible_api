#!/usr/bin/env bash
# Create a fresh KJV OpenSearch index with the case-preserving text subfield.
# This deliberately never deletes or overwrites an existing index.
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/create_case_sensitive_index.sh --index NAME [--url URL] [--db PATH]

Creates and fills a brand-new index using scripts/opensearch_kjv_mapping.json.
The target index must not already exist. Point OPENSEARCH_INDEX at the new
index only after verification; retain the prior index for rollback.
EOF
}

index=""
url="http://localhost:9200"
db="data/kjv.db"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --index) index="${2:-}"; shift 2 ;;
    --url) url="${2:-}"; shift 2 ;;
    --db) db="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ -z "$index" ]]; then
  echo "--index is required" >&2
  usage >&2
  exit 2
fi
if [[ ! "$index" =~ ^[a-zA-Z0-9._-]+$ ]]; then
  echo "Index name contains unsupported characters: $index" >&2
  exit 2
fi

status="$(curl -sS -o /dev/null -w '%{http_code}' "$url/$index")"
case "$status" in
  404) ;;
  200) echo "Refusing to overwrite existing index '$index'. Choose a new versioned name." >&2; exit 1 ;;
  *) echo "Could not check index '$index' (HTTP $status)." >&2; exit 1 ;;
esac

python3 scripts/index_kjv_to_opensearch.py --url "$url" --index "$index" --db "$db"

mapping="$(curl -fsS "$url/$index/_mapping/field/text.case_sensitive")"
if [[ "$mapping" != *'case_sensitive'* ]]; then
  echo "Index was created, but text.case_sensitive was not found in its mapping." >&2
  exit 1
fi

echo "Created and populated $index. Set OPENSEARCH_INDEX=$index only after smoke testing."
