# Bible API

A Go REST API and React UI for reading and searching the King James Bible.

## Quick URLs

Local (Docker Compose on the API box):

- API base: `http://localhost:8000`
- Web UI: `http://localhost:8000/v2`
- Book listing: `http://localhost:8000/bible/list_books`
- Search: `http://localhost:8000/bible/search?q=grace`
- Autocomplete: `http://localhost:8000/bible/suggest?q=gra`
- Random verse: `http://localhost:8000/bible/random_verse`
- API docs: `http://localhost:8000/docs`
- OpenSearch: `http://localhost:9200`
- OpenSearch Dashboards: `http://localhost:5601`

Public (nginx path prefix on prsmusa.com):

- UI: https://prsmusa.com/bible/v2
- Preview image: https://prsmusa.com/bible/v2/og-image.png
- Docs: https://prsmusa.com/docs
- Landing link: https://prsmusa.com/

> If search returns `lookup opensearch on 127.0.0.11:53: no such host`, the API container can't resolve the `opensearch` service name. Start all services via `docker compose` so they share the same network.

> OpenSearch 2.12+ requires an initial admin password. Set `OPENSEARCH_INITIAL_ADMIN_PASSWORD` in `.env` before `docker compose up`. Security is disabled in this compose file; the variable is still required by the image.

- A raw high performance RESTful API written in Go
- King James Version Pure Cambridge Text
- No ads, No distractions, not ever.
- Hamburger navigation menu on every page (Books, Search, Docs, Donations, Settings, cross-link to v2/classic)
- Font settings: choose from Default, Blackletter (Gothic), Renaissance, or Classic Serif — persists via localStorage
- All Bible text preloaded into memory at startup for instant reads (zero OpenSearch latency for chapter/verse/random)
- Easy navigation
  - [Simple book listing](https://prsmusa.com/bible/list_books)
  - [Random Verse](https://prsmusa.com/bible/random_verse)
  - [JSON output](https://prsmusa.com/bible/random_verse?json=true) (`?json=true`)
  - Forward / previous chapter, books list, [verse ranges](https://prsmusa.com/bible/EPHESIANS/2/8-9)
  - Search with per-book chart: `https://prsmusa.com/bible/search?q=heart`
  - Predictive search bar on every page (autocomplete via OpenSearch)

## Features

### Bible Content

- Complete King James Version
- Book navigation with clickable chapters
- Verse range support (e.g., `/bible/ROMANS/5/1-5`)
- Full-text search powered by OpenSearch with per-book match chart
- Predictive autocomplete search bar on every page
- Hamburger navigation menu on every page (Books, Search, Docs, Settings, cross-link between classic/v2)
- Font settings with 4 choices: Default, Blackletter (Gothic), Renaissance, Classic Serif — saved in localStorage, persists across pages and sessions
- All ~31k verses preloaded into memory at startup — chapter reads, verse lookups, and random verse are served from cache with zero network latency

### Architecture

- **OpenSearch** — Bible content search and autocomplete (`kjv_v2` index). Chapter/verse/random reads are served from an in-memory preload after startup.
- **Redis** — rate limiting, session storage, notes safety gate
- **Postgres** — optional 3-year activity audit (`DATABASE_URL`). Empty URL disables durable analytics; Prometheus `/metrics` still works.
- **React UI** — Vite app in `web/`, baked into the Docker image at `web/dist` and served at `/v2`

### Donations (Stripe Checkout)

The Donations menu item opens `/donate`, where supporters can choose a one-time or
monthly USD donation and are redirected to Stripe Checkout. Configure these values
outside source control before enabling payments:

```bash
STRIPE_SECRET_KEY=replace_with_rotated_secret
PUBLIC_BASE_URL=https://prsmusa.com
```

`STRIPE_SECRET_KEY` is server-only. Do not expose it in browser code, commit it to
the repository, or put it in a Vite environment variable. Monthly donations use
Stripe Billing; one-time donations request Stripe invoice creation. Configure Stripe
Tax and your account's donation receipt/tax settings in the Stripe Dashboard.

### Rate Limiting

- **5 requests per second** per IP address
- **1-minute blocking** when limit exceeded
- Redis-based storage with automatic TTL cleanup
- Rate limit headers in all responses
- Admin endpoints for manual IP management

## Quick Start

### Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/r4wm/bible_api.git
cd bible_api

# Set the required OpenSearch admin password
export OPENSEARCH_INITIAL_ADMIN_PASSWORD=YourStrongPassword123!

# Copy env template and fill OPENSEARCH_INITIAL_ADMIN_PASSWORD
cp .env.example .env

# Start OpenSearch, Redis, and Bible API (Postgres is an optional profile)
docker compose up -d

# Verify the service is running
curl "http://localhost:8000/health"
```

The bible_api service waits for OpenSearch to be healthy before starting. However, **Bible content endpoints (`/bible/*`) will return empty results until you index the data**. On first run:

```bash
# Index the KJV data into OpenSearch (required before Bible endpoints work)
python3 scripts/index_kjv_to_opensearch.py --url http://localhost:9200

# Verify the index is working
curl "http://localhost:8000/bible/GENESIS/1?json=true"
```

### Reindexing OpenSearch

If autocomplete (`/bible/suggest`) returns poor results after upgrading, or if you've updated the indexer script, delete and rebuild the index:

```bash
curl -X DELETE http://localhost:9200/kjv_v2
python3 scripts/index_kjv_to_opensearch.py --url http://localhost:9200
```

This is required whenever `text_suggest` indexing logic changes in `scripts/index_kjv_to_opensearch.py`.

### Case-sensitive v2 search rollout

`GET /bible/v2/search` accepts two optional controls:

- `match=any` (default), `match=all`, or `match=phrase`
- `case_sensitive=true|false` (default `false`)

`case_sensitive=true` queries the `text.case_sensitive` field. That field is populated only when a verse is indexed with the current mapping, so **do not enable or rely on case-sensitive search against an existing index until it has been rebuilt.**

Use a new versioned index rather than deleting the live index in place. The helper below refuses to overwrite an existing index:

```bash
# Create and populate a new index using the current mapping.
scripts/create_case_sensitive_index.sh \
  --url http://localhost:9200 \
  --index kjv_v2_case_sensitive_20260809

# Point the API at the new index, restart it, then verify the behavior.
export OPENSEARCH_INDEX=kjv_v2_case_sensitive_20260809
curl 'http://localhost:8000/bible/v2/search?q=God&case_sensitive=true'
curl 'http://localhost:8000/bible/v2/search?q=god&case_sensitive=true'
```

Keep the previous index until the new one has passed API and UI smoke tests. The case-preserving analyzer uses normal token matching: `match=phrase` means an exact sequence of tokens, not byte-for-byte punctuation or whitespace matching.

### Manual Setup

```bash
# Ensure OpenSearch is running at localhost:9200
# Ensure Redis is running at localhost:6379

# Build
go build -o bible_api cmd/bible_api.go

# Run
./bible_api
```

## Production Topology

The public site at **prsmusa.com** is an nginx edge. The API, UI, Redis, and OpenSearch all run in Docker on a home host (`baser4wm@10.0.0.68`). A reverse SSH tunnel publishes the API onto loopback on the edge — the Bible process is not installed on the Linode.

```
  Internet
     │
     ▼
  nginx :443  (r4wm@prsmusa.com / Linode)
     │  TLS + path routing
     │  /etc/nginx/sites-enabled/prsmusa.com
     │  snippets: prsmusa-bible-user-api.conf, prsmusa-bible-legacy-prefix.conf
     │
     ├─ /                    static landing page  /var/www/prsmusa.com
     ├─ /bible/v2            proxy → 127.0.0.1:18000/v2   (sub_filter rewrites "/v2/" → "/bible/v2/")
     ├─ /bible/v2/search     proxy → 127.0.0.1:18000      (must stay more specific than the UI prefix)
     ├─ /bible/, /auth/, /user/, /donate, /docs
     │                       proxy → 127.0.0.1:18000
     └─ /v2                  308 → /bible/v2
            │
            ▼
  ssh -R 127.0.0.1:18000:127.0.0.1:8000   r4wm@172.236.115.113
  (started on the API box by crontab @reboot → ~/bin/prsmusa-bible-tunnel)
            │
            ▼
  Docker on 10.0.0.68
     bible_api :8000   (image built from this repo's Dockerfile)
     redis :6379
     opensearch :9200
     opensearch_dashboards :5601
```

### Edge (prsmusa.com)

- nginx terminates Let's Encrypt TLS and path-routes Bible traffic to **loopback :18000**, not to a local binary.
- The Vite app is built with `base: "/v2/"`. nginx `sub_filter` rewrites those asset URLs to `/bible/v2/` for the public prefix.
- OG / Twitter tags in `web/index.html` use **absolute** `https://prsmusa.com/bible/v2/...` URLs. Scrapers do not run JavaScript and do not resolve relative paths. The card image is `web/public/og-image.png` (regenerate with `python3 scripts/make_og_image.py`).
- There is no `bible_api.service`, no `/opt/bible_api`, and no `/usr/local/bin/bible_api` on this host. Do not rsync a Go binary here.

### API box (10.0.0.68)

- Repo: `~/github/bible_api`
- `docker compose` project `bible_api`. Container name `bible_api` publishes `0.0.0.0:8000`.
- Dockerfile builds the React UI (`web/`) then the Go binary. The running server reads `web/dist` from disk (`kjv/ui.go`).
- Reverse tunnel: `~/bin/prsmusa-bible-tunnel` (`ssh -R 127.0.0.1:18000:127.0.0.1:8000`). Keepalive is in the ssh options; crontab restarts it at boot with `flock`.
- `.env` is compose interpolation only (`STRIPE_SECRET_KEY`, `PUBLIC_BASE_URL`, `OPENSEARCH_INITIAL_ADMIN_PASSWORD`). Do not commit it.

### Deploy (this is the live path)

Run **on 10.0.0.68** from the repo root:

```bash
make deploy          # tag rollback image, docker compose up -d --build --no-deps bible_api, smoke-test
make deploy-check    # curl local /health and https://prsmusa.com/bible/v2 (OG tags + image)
```

`scripts/deploy.sh` leaves Redis and OpenSearch running. It does not start Postgres. After a successful rebuild the previous image is tagged `bible_api-bible_api:rollback-<UTC>`.

A hot-swap of `web/dist` inside the running container is possible (`kjv/ui.go` serves files from disk) but is **not durable**. The next `make deploy` or `docker compose up --build` replaces the container from the image. Commit UI changes and rebuild.

### Optional Postgres analytics

Durable login/chapter/search history is behind compose profile `analytics`. The live public container currently runs with `DATABASE_URL` unset (Prometheus `/metrics` only).

```bash
# .env
POSTGRES_PASSWORD=pick-a-secret
DATABASE_URL=postgresql://bible_api:pick-a-secret@postgres:5432/bible_api?sslmode=disable

docker compose --profile analytics up -d postgres
docker compose up -d --no-deps bible_api
```

### Verifying a deploy

```bash
# on 10.0.0.68
curl -sf http://127.0.0.1:8000/health
docker compose logs --tail 50 bible_api
pgrep -af 'ssh .*18000:127.0.0.1:8000'    # reverse tunnel

# from anywhere
curl -sI https://prsmusa.com/bible/v2
curl -s https://prsmusa.com/bible/v2 | grep og:image
curl -sI https://prsmusa.com/bible/v2/og-image.png     # Content-Type: image/png
curl -s 'https://prsmusa.com/bible/random_verse?json=true' | head
```

If the public UI 502s but `curl localhost:8000/health` works on the API box, the reverse tunnel is down — restart `~/bin/prsmusa-bible-tunnel`. If verse endpoints are empty, OpenSearch is down or the `kjv_v2` index was not loaded.

## Rate Limiting

The API includes built-in rate limiting to prevent abuse:

- **Limit**: 5 requests per second per IP
- **Block Duration**: 1 minute when exceeded
- **Headers**: Rate limit info in response headers

### Rate Limit Headers

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset: 1640995200
```

### Admin Endpoints

```bash
# Check IP status
GET /admin/rate-limit/{ip}

# Block IP manually
POST /admin/block-ip
{"ip": "192.168.1.100", "duration": "10m"}

# Unblock IP
DELETE /admin/unblock-ip/{ip}
```

## Testing

```bash
# Test random verse
curl "http://localhost:8000/bible/random_verse?json=true"

# Test search
curl "http://localhost:8000/bible/search?q=love&json=true"

# Test verse range
curl "http://localhost:8000/bible/JOHN/3/16-17?json=true"

# Test autocomplete
curl "http://localhost:8000/bible/suggest?q=grace"

# Test old English font mode (browser)
# http://localhost:8000/bible/GENESIS/1?olde=true

# Test rate limiting
./test_rate_limit.sh
./test_logging.sh
```

To use the public version, visit [https://prsmusa.com/bible/v2](https://prsmusa.com/bible/v2).

## API Endpoints

### Bible Content

- `GET /bible/list_books` - List all Bible books (66 books in canonical order)
- `GET /bible/{book}` - List chapters in a book
- `GET /bible/{book}/{chapter}` - Get all verses in a chapter
- `GET /bible/{book}/{chapter}/{verse}` - Get specific verse
- `GET /bible/{book}/{chapter}/{start-end}` - Get verse range
- `GET /bible/search?q={query}` - Full-text search with per-book chart
- `GET /bible/suggest?q={prefix}` - Autocomplete suggestions (n-gram completion)
- `GET /bible/random_verse` - Get random verse

### v2 (OpenSearch Direct)

- `GET /bible/v2/search?q={query}&match={any|all|phrase}&case_sensitive={true|false}` - OpenSearch full-text search (JSON). Defaults: `match=any`, `case_sensitive=false`.
- `GET /bible/v2/suggest?q={prefix}` - Verse prefix suggestions (JSON)
- `PUT /bible/v2/synonyms/{set}` - Replace synonym set (JWT required)
- `POST /bible/v2/synonyms/{set}` - Append to synonym set (JWT required)
- `DELETE /bible/v2/synonyms/{set}` - Remove from synonym set (JWT required)
- `GET /v2` - Web UI (Google login + search/suggest)
- `GET /docs` - API documentation (HTML)
- `GET /docs.json` - API documentation (JSON)

### Auth

- `POST /auth/google/token` - Exchange Google ID token for app JWT + session cookie
- `GET /auth/config` - Auth configuration (Google client id)
- `GET /auth/me` - Get current session token
- `POST /auth/logout` - Clear session
- `POST /admin/token` - Mint token with `X-Internal-Secret` (internal use)

### Admin (Rate Limiting)

- `GET /admin/rate-limit/{ip}` - Check IP rate limit status
- `POST /admin/block-ip` - Manually block an IP
- `DELETE /admin/unblock-ip/{ip}` - Unblock an IP

### System

- `GET /health` - Health check (bypasses rate limiting)

### URL Parameters

- `?json=true` - Return JSON response instead of HTML
- `?show_italics=true` - Show italicized text in Bible verses (where applicable)
- `?n={limit}` - Limit number of results. Search defaults to 10000. Suggest defaults to 20 (max 50).

Font selection is handled via the **Settings** menu (hamburger menu → Settings) and saved in the browser's localStorage — no URL parameter needed.

**Examples:**

```bash
# Get verse in JSON format
curl "http://localhost:8000/bible/JOHN/3/16?json=true"

# Get verse with italics shown
curl "http://localhost:8000/bible/PSALMS/23/1?show_italics=true"

# Combine parameters
curl "http://localhost:8000/bible/ROMANS/8/28?json=true&show_italics=true"
```

## Development

### Prerequisites

- Go 1.20+
- Redis server
- OpenSearch 2.x (required for all Bible content)

### Build from Source

```bash
# Install dependencies
go mod tidy

# Build
go build -o bible_api cmd/bible_api.go

# Run (requires OpenSearch and Redis to be running)
./bible_api
```

### UI (React)

```bash
cd web
npm install
npm run dev
```

Build static UI for `/v2`:

```bash
cd web
npm run build
```

The Go server serves `web/dist` at `/v2`. Docker builds the UI during image build.

Link-preview tags live in `web/index.html` (Open Graph + Twitter). The card image is `web/public/og-image.png`. Vite copies `public/` into `dist/` unchanged. After changing the card art:

```bash
python3 scripts/make_og_image.py
make deploy
```

### Environment Variables

```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
OPENSEARCH_URL=http://localhost:9200
OPENSEARCH_INDEX=kjv_v2
OPENSEARCH_SYNONYMS_SET=kjv_synonyms
JWT_SECRET=change_me
JWT_ISSUER=bible_api
JWT_AUDIENCE=bible_api_clients
JWT_TTL_SECONDS=3600
SESSION_TTL_SECONDS=3600
SESSION_COOKIE_NAME=bible_api_session
SESSION_COOKIE_SECURE=false
INTERNAL_TOKEN_SECRET=change_me
GOOGLE_CLIENT_ID=your-google-client-id
PUBLIC_BASE_URL=https://prsmusa.com
STRIPE_SECRET_KEY=
DATABASE_URL=   # empty = no Postgres audit trail
```

## Legacy Text Sources

Legacy English plain-text bible files are stored in `assets/texts/`, one translation per directory.

- Source: [BibleSuper SourceForge - All Bibles (Plain Text) EN-English](https://sourceforge.net/projects/biblesuper/files/All%20Bibles%20-%20Plain%20Text/EN-English/?utm_source=chatgpt.com)
- `kjv.txt` is intentionally excluded from this set for now and should be skipped in import workflows.

### File Abbreviations

- `asv.txt`: American Standard Version
- `asvs.txt`: American Standard Version (with Strong's numbers)
- `bishops.txt`: Bishops' Bible
- `coverdale.txt`: Coverdale Bible
- `geneva.txt`: Geneva Bible
- `kjv_strongs.txt`: King James Version (with Strong's numbers)
- `net.txt`: New English Translation
- `tyndale.txt`: Tyndale Bible
- `web.txt`: World English Bible

### Integrity Check (MD5)

Each translation directory includes its own `MD5SUM` file.

```bash
cd assets/texts/asv
md5sum -c MD5SUM
```

### Per-Translation Layout

Each translation directory (example: `assets/texts/asv/`) contains:

- `<translation>.txt` source text
- `README.md` with source and notes
- `CHANGELOG.md` for change history
- `MD5SUM` for integrity verification
- `scripts/import_to_sqlite.py` for SQLite ingestion
- `scripts/import_to_opensearch.py` for OpenSearch ingestion
- `scripts/verify_import.py` for checksum and import checks

## Project Structure

```
bible_api/
├── cmd/
│   └── bible_api.go          # Main application entry point
├── kjv/
│   ├── kjv.go               # Core Bible API handlers (OpenSearch-backed)
│   ├── v2.go                # OpenSearch helpers, v2 endpoints, suggest
│   ├── admin.go             # Admin endpoints for rate limiting
│   ├── auth.go              # Google OAuth + JWT auth
│   ├── ui.go                # Static UI serving
│   └── templates.go         # HTML templates with search bar
├── middleware/
│   └── rate_limiter.go      # Redis-based rate limiting
├── assets/
│   └── texts/               # Legacy plain-text Bible translations
├── web/                      # React UI (served at /v2, public OG image in web/public/)
├── scripts/
│   ├── deploy.sh                   # Rebuild bible_api container on this host
│   ├── make_og_image.py            # Render web/public/og-image.png
│   └── index_kjv_to_opensearch.py  # Index Bible data into OpenSearch
├── docker-compose.yml        # OpenSearch + Redis + Bible API (+ optional postgres profile)
├── Dockerfile
├── test_rate_limit.sh
├── test_logging.sh
└── RATE_LIMITING.md
```

## Documentation

- [Rate Limiting Guide](RATE_LIMITING.md) - Detailed rate limiting documentation
- [API Examples](#testing) - Example API calls and responses

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source. The King James Version text is in the public domain.
