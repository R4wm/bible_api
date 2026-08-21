#!/usr/bin/env bash
# Rebuild the bible_api container on this host (the data plane).
#
# Public traffic never hits this host directly. nginx on prsmusa.com proxies
# /bible, /v2, /auth, /user, /donate, /docs to 127.0.0.1:18000, which is a
# reverse SSH tunnel to this machine's :8000:
#
#   Internet -> nginx :443 on prsmusa.com
#            -> 127.0.0.1:18000  (ssh -R from this host)
#            -> 127.0.0.1:8000   (docker container bible_api)
#
# Run this script ON the API box (baser4wm@10.0.0.68), from the repo root, or
# via `make deploy`. Do not rsync a binary to prsmusa.com — there is no
# bible_api.service there.
#
#   ./scripts/deploy.sh           # tag rollback, rebuild, recreate bible_api
#   ./scripts/deploy.sh --check   # smoke-test local + public URLs only
#
# Redis and OpenSearch are left running (--no-deps). Postgres is optional and
# is not started unless you `docker compose --profile analytics up`.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

CHECK_ONLY=0
while [ $# -gt 0 ]; do
    case "$1" in
        --check) CHECK_ONLY=1; shift ;;
        -h|--help)
            sed -n '2,22p' "$0" | sed 's/^# \?//'
            exit 0
            ;;
        *) echo "unknown arg: $1" >&2; exit 1 ;;
    esac
done

log()  { printf '\033[1;36m[deploy]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[deploy] WARN:\033[0m %s\n' "$*" >&2; }

smoke() {
    local url="$1"
    local code
    code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "$url")"
    if [ "$code" != "200" ]; then
        echo "expected 200 from $url, got $code" >&2
        return 1
    fi
    log "ok $code $url"
}

check() {
    log "local health"
    smoke "http://127.0.0.1:8000/health"
    smoke "http://127.0.0.1:8000/v2"
    log "public edge (nginx + tunnel)"
    smoke "https://prsmusa.com/bible/v2"
    smoke "https://prsmusa.com/bible/v2/og-image.png"
    if ! curl -sS --max-time 15 "https://prsmusa.com/bible/v2" | grep -q 'property="og:image"'; then
        echo "public /bible/v2 is missing og:image" >&2
        return 1
    fi
    if ! curl -sS -I --max-time 15 "https://prsmusa.com/bible/v2/og-image.png" | grep -qi 'content-type: image/png'; then
        echo "public og-image.png is not image/png" >&2
        return 1
    fi
    log "og tags and preview image look good"
}

if [ "$CHECK_ONLY" = "1" ]; then
    check
    exit 0
fi

if ! docker compose config >/dev/null; then
    echo "docker compose config failed. Check .env (OPENSEARCH_INITIAL_ADMIN_PASSWORD, optional POSTGRES_*)." >&2
    exit 1
fi

if ! docker compose ps --status running --format '{{.Name}}' | grep -qx bible_api_redis; then
    warn "bible_api_redis is not running; starting redis + opensearch first"
    docker compose up -d redis opensearch
fi

stamp="$(date -u +%Y%m%d-%H%M%S)"
if docker image inspect bible_api-bible_api >/dev/null 2>&1; then
    log "tag current image as bible_api-bible_api:rollback-$stamp"
    docker tag bible_api-bible_api "bible_api-bible_api:rollback-$stamp"
fi

log "rebuild and recreate bible_api (--no-deps keeps redis/opensearch)"
docker compose up -d --build --no-deps bible_api

log "wait for /health"
for i in $(seq 1 30); do
    if curl -sf --max-time 2 "http://127.0.0.1:8000/health" >/dev/null; then
        break
    fi
    if [ "$i" = "30" ]; then
        echo "bible_api did not become healthy" >&2
        docker compose logs --tail 80 bible_api >&2
        exit 1
    fi
    sleep 2
done

if ! ss -ltn | grep -q ':8000'; then
    warn "nothing is listening on :8000"
fi

if ! pgrep -af '/usr/bin/ssh .*18000:127.0.0.1:8000' >/dev/null; then
    warn "reverse tunnel to prsmusa.com:18000 is not running"
    warn "start it with: ~/bin/prsmusa-bible-tunnel  (crontab @reboot)"
fi

check
log "done. rollback image: bible_api-bible_api:rollback-$stamp"
log "logs: docker compose logs -f bible_api"
