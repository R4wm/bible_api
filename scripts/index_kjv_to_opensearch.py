#!/usr/bin/env python3
import argparse
import json
import sqlite3
import sys
import urllib.request

DEFAULT_INDEX = "bible_kjv_v1"


def http_request(method, url, body=None, headers=None):
    req = urllib.request.Request(url, data=body, method=method)
    if headers:
        for k, v in headers.items():
            req.add_header(k, v)
    try:
        with urllib.request.urlopen(req) as resp:
            return resp.getcode(), resp.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()


def ensure_index(base_url, index_name):
    mapping = {
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 0
        },
        "mappings": {
            "properties": {
                "book": {"type": "keyword"},
                "chapter": {"type": "integer"},
                "verse": {"type": "integer"},
                "text": {"type": "text"},
                "ordinal_verse": {"type": "integer"},
                "ordinal_book": {"type": "integer"},
                "testament": {"type": "keyword"}
            }
        }
    }

    code, body = http_request(
        "PUT",
        f"{base_url}/{index_name}",
        body=json.dumps(mapping).encode("utf-8"),
        headers={"Content-Type": "application/json"},
    )

    if code in (200, 201):
        return True

    if code == 400 and b"resource_already_exists_exception" in body:
        return True

    sys.stderr.write(f"Failed to create index {index_name}: {code} {body.decode('utf-8', 'ignore')}\n")
    return False


def bulk_index(base_url, index_name, rows, batch_size):
    count = 0
    batch = []

    def flush():
        if not batch:
            return
        payload = "\n".join(batch) + "\n"
        code, body = http_request(
            "POST",
            f"{base_url}/_bulk",
            body=payload.encode("utf-8"),
            headers={"Content-Type": "application/x-ndjson"},
        )
        if code not in (200, 201):
            sys.stderr.write(f"Bulk index failed: {code} {body.decode('utf-8', 'ignore')}\n")
            sys.exit(1)
        batch.clear()

    for row in rows:
        doc = {
            "book": row[0],
            "chapter": row[1],
            "verse": row[2],
            "text": row[3],
            "ordinal_verse": row[4],
            "ordinal_book": row[5],
            "testament": row[6],
        }
        doc_id = row[4]
        batch.append(json.dumps({"index": {"_index": index_name, "_id": doc_id}}))
        batch.append(json.dumps(doc, ensure_ascii=False))
        count += 1
        if count % batch_size == 0:
            flush()
            print(f"Indexed {count} verses...")

    flush()
    print(f"Indexed {count} verses total.")


def main():
    parser = argparse.ArgumentParser(description="Index KJV SQLite DB into OpenSearch")
    parser.add_argument("--db", default="data/kjv.db", help="Path to kjv SQLite database")
    parser.add_argument("--index", default=DEFAULT_INDEX, help="OpenSearch index name")
    parser.add_argument("--url", default="http://localhost:9200", help="OpenSearch base URL")
    parser.add_argument("--batch", type=int, default=1000, help="Bulk batch size")
    args = parser.parse_args()

    if not ensure_index(args.url, args.index):
        sys.exit(1)

    conn = sqlite3.connect(args.db)
    try:
        cursor = conn.execute(
            "select book, chapter, verse, text, ordinal_verse, ordinal_book, testament "
            "from kjv order by ordinal_verse"
        )
        bulk_index(args.url, args.index, cursor, args.batch)
    finally:
        conn.close()


if __name__ == "__main__":
    main()
