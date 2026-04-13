# Bible API

## Quick URLs (Local)

- API base: `http://localhost:8000`
- Book listing: `http://localhost:8000/bible/list_books`
- Search: `http://localhost:8000/bible/search?q=grace`
- Autocomplete: `http://localhost:8000/bible/suggest?q=gra`
- Random verse: `http://localhost:8000/bible/random_verse`
- Web UI: `http://localhost:8000/v2`
- OpenSearch: `http://localhost:9200`
- OpenSearch Dashboards: `http://localhost:5601`
- API docs: `http://localhost:8000/docs`

> If search returns `lookup opensearch on 127.0.0.11:53: no such host`, the API container can't resolve the `opensearch` service name. Start all services via `docker compose` so they share the same network.

> OpenSearch 2.12+ requires an initial admin password. Set `OPENSEARCH_INITIAL_ADMIN_PASSWORD` in your environment before running `docker compose up`.

- A raw high performance RESTful API written in Go
- King James Version Pure Cambridge Text
- No ads, No distractions, not ever.
- Hamburger navigation menu on every page (Books, Search, Docs, Settings, cross-link to v2/classic)
- Font settings: choose from Default, Blackletter (Gothic), Renaissance, or Classic Serif — persists via localStorage
- All Bible text preloaded into memory at startup for instant reads (zero OpenSearch latency for chapter/verse/random)
- Easy navigation
  - [Simple book listing and buttons choice](https://mintz5.duckdns.org/bible/list_books)
  - [Random Verse Generator](https://mintz5.duckdns.org/bible/random_verse)
  - [All pages support json output](https://mintz5.duckdns.org/bible/random_verse?json=true)
    - provide argument: `?json=true`
  - Forward chapter button (if applicable)
  - Previous chapter button (if applicable)
  - Books link button in Chapter selection
  - [Supports verse ranges](https://mintz5.duckdns.org/bible/EPHESIANS/2/8-9)
  - Search feature with per-book chart visualization
    - Example: `https://mintz5.duckdns.org/bible/search?q=heart`
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

- **OpenSearch** — all Bible content reads, full-text search, autocomplete suggestions
- **Redis** — rate limiting, session storage
- No SQLite dependency. OpenSearch is the sole data source for Bible content.

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

# Start all services (OpenSearch, Redis, Bible API)
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

### Manual Setup

```bash
# Ensure OpenSearch is running at localhost:9200
# Ensure Redis is running at localhost:6379

# Build
go build -o bible_api cmd/bible_api.go

# Run
./bible_api
```

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

To use the public version, visit the [bible_api](https://mintz5.duckdns.org/bible/list_books)

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

- `GET /bible/v2/search?q={query}` - OpenSearch full-text search (JSON)
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

The Go server serves `web/dist` at `/v2`.
Docker builds the UI automatically during image build.

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
├── web/                      # React UI (served at /v2)
├── scripts/
│   └── index_kjv_to_opensearch.py  # Index Bible data into OpenSearch
├── docker-compose.yml        # OpenSearch + Redis + Bible API
├── Dockerfile
├── start_with_redis.sh
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
