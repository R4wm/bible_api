#!/usr/bin/env bash
# Build and deploy bible_api to prsmusa.com.
#
# By default builds the frontend (web/dist) AND the Go binary, rsyncs both,
# then restarts bible_api.service. Flags:
#
#   --ui-only       skip the Go build/rsync (fast iteration on UI changes)
#   --backend-only  skip the frontend build/rsync
#   --no-restart    do not call systemctl restart (handy for dry runs)
#   --host HOST     override target host (default: r4wm@prsmusa.com)
#
# The Go binary is cross-compiled with CGO_ENABLED=0 so it's a static linux
# amd64 ELF, which sidesteps glibc version drift between dev box and target.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

DO_UI=1
DO_BE=1
DO_RESTART=1
TARGET="${BIBLE_API_DEPLOY_HOST:-r4wm@prsmusa.com}"

while [ $# -gt 0 ]; do
    case "$1" in
        --ui-only)      DO_BE=0; shift ;;
        --backend-only) DO_UI=0; shift ;;
        --no-restart)   DO_RESTART=0; shift ;;
        --host)         TARGET="$2"; shift 2 ;;
        -h|--help)
            sed -n '2,15p' "$0" | sed 's/^# \?//'
            exit 0
            ;;
        *) echo "unknown arg: $1" >&2; exit 1 ;;
    esac
done

log()  { printf '\033[1;36m[deploy]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[deploy] WARN:\033[0m %s\n' "$*" >&2; }

# --- build ---

if [ "$DO_UI" = "1" ]; then
    log "build frontend (web/)"
    cd "$REPO_ROOT/web"
    if [ ! -d node_modules ] || [ package.json -nt node_modules ]; then
        npm install
    fi
    npm run build
    cd "$REPO_ROOT"
    [ -f web/dist/index.html ] || { echo "web/dist/index.html missing after build" >&2; exit 1; }
fi

if [ "$DO_BE" = "1" ]; then
    log "build backend (CGO_ENABLED=0 linux/amd64)"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -ldflags='-s -w' -o bible_api ./cmd/bible_api.go
    file bible_api | grep -q 'ELF 64-bit.*x86-64' \
        || { echo "built binary not linux x86-64" >&2; exit 1; }
fi

# --- ship ---

if [ "$DO_UI" = "1" ]; then
    log "rsync web/dist -> ${TARGET}:/opt/bible_api/web/dist/"
    ssh "$TARGET" 'sudo mkdir -p /opt/bible_api/web && sudo chown $USER /opt/bible_api/web'
    rsync -az --delete web/dist/ "${TARGET}:/opt/bible_api/web/dist/"
fi

if [ "$DO_BE" = "1" ]; then
    # rsync to tmp, then sudo-install in place. Avoids "rsync as root" and
    # keeps the binary atomic (install does the rename).
    log "rsync bible_api binary -> ${TARGET}:/tmp/bible_api.new"
    rsync -az bible_api "${TARGET}:/tmp/bible_api.new"
    log "sudo install -> /usr/local/bin/bible_api"
    ssh -t "$TARGET" 'sudo install -m 755 /tmp/bible_api.new /usr/local/bin/bible_api && rm -f /tmp/bible_api.new'
fi

# --- restart ---

if [ "$DO_RESTART" = "1" ]; then
    log "sudo systemctl restart bible_api.service"
    ssh -t "$TARGET" 'sudo systemctl restart bible_api.service && sudo systemctl is-active bible_api.service'
else
    warn "skipping systemctl restart (--no-restart set)"
fi

log "done. tail logs with:  ssh ${TARGET} 'journalctl -u bible_api -f'"
