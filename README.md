# Bible API

## Quick URLs (Local)

- API base: `http://localhost:8000`
- Search endpoint (example): `http://localhost:8000/bible/v2/search?q=genesis`
- Web UI: `http://localhost:8000/v2`
- OpenSearch: `http://localhost:9200`
- OpenSearch Dashboards: `http://localhost:5601`

> If search returns `lookup opensearch on 127.0.0.11:53: no such host`, the API container can’t resolve the `opensearch` service name. Start all services via `docker compose` so they share the same network.

> OpenSearch 2.12+ requires an initial admin password. Set `OPENSEARCH_INITIAL_ADMIN_PASSWORD` in your environment before running `docker compose up`.

- A raw high performance RESTful API written in Go
- King James Version Pure Cambridge Text
- No ads, No distractions, not ever.
- **NEW**: Redis-based rate limiting for abuse prevention
- Easy navigation
  - [Simple book listing and buttons choice](https://mintz5.duckdns.org/bible/list_books)
  - [Random Verse Generator](https://mintz5.duckdns.org/bible/random_verse)
  - [All pages support json output](https://mintz5.duckdns.org/bible/random_verse?json=true)
    - provide argument: `?json=true`
  - Forward chapter button (if applicable)
  - Previous chapter button (if applicable)
  - Books link button in Chapter selection
  - [Supports verse ranges](https://mintz5.duckdns.org/bible/EPHESIANS/2/8-9)
  - Search feature
    - Example: `https://mintz5.duckdns.org/bible/search?q=heart`

## ⚡ Features

### 📖 Bible Content

- Complete King James Version
- Book navigation with clickable chapters
- Verse range support (e.g., `/bible/romans/5/1-5`)
- Full-text search across all books

### 🛡️ Rate Limiting (NEW)

- **5 requests per second** per IP address
- **1-minute blocking** when limit exceeded
- Redis-based storage with automatic TTL cleanup
- Rate limit headers in all responses
- Admin endpoints for manual IP management
- Comprehensive logging of rate limit events

### 🔧 Technical Features

- JSON and HTML response formats
- RESTful API design
- SQLite database backend
- Redis for rate limiting
- Docker support with docker-compose
- Graceful error handling

## 🚀 Quick Start

### Method 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/r4wm/bible_api.git
cd bible_api

# Start with Redis and Bible API
docker-compose up -d

# API will be available at http://localhost:8000
curl "http://localhost:8000/health"
```

### OpenSearch (Single Node)

```bash
# Start OpenSearch and Dashboards
make opensearch-up

# Optional on some Linux hosts
sudo sysctl -w vm.max_map_count=262144

# Index the KJV DB into OpenSearch
make index-kjv
```

OpenSearch runs at `http://localhost:9200` and Dashboards at `http://localhost:5601`.

### Reindexing OpenSearch

If autocomplete (`/bible/suggest`) returns poor results after upgrading, delete and rebuild the OpenSearch index:

```bash
curl -X DELETE http://localhost:9200/kjv_v2
python3 scripts/index_kjv_to_opensearch.py --db /data/kjv.db
```

This is required whenever `text_suggest` indexing logic changes in `scripts/index_kjv_to_opensearch.py`.

### Method 2: Manual Setup

```bash
# Start Redis
redis-server

# Or use the convenience script
./start_with_redis.sh
```

## 📊 Rate Limiting

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

## 🧪 Testing

### Test Rate Limiting

```bash
./test_rate_limit.sh      # Test rate limiting behavior
./test_logging.sh         # Test with logging output
```

### Manual Testing

```bash
# Test random verse
curl "http://localhost:8000/bible/random_verse?json=true"

# Test search
curl "http://localhost:8000/bible/search?q=love&json=true"

# Test verse range
curl "http://localhost:8000/bible/JOHN/3/16-17?json=true"
```

To use public version of running API, visit the [bible_api](https://mintz5.duckdns.org/bible/list_books)

## 📋 API Endpoints

### Bible Content

- `GET /bible/list_books` - List all Bible books
- `GET /bible/{book}` - List chapters in a book
- `GET /bible/{book}/{chapter}` - Get all verses in a chapter
- `GET /bible/{book}/{chapter}/{verse}` - Get specific verse
- `GET /bible/{book}/{chapter}/{start-end}` - Get verse range
- `GET /bible/search?q={query}` - Search Bible text
- `GET /bible/suggest?q={prefix}` - Autocomplete suggestions (OpenSearch)
- `GET /bible/random_verse` - Get random verse

### v2 (OpenSearch)

- `GET /bible/v2/search?q={query}` - OpenSearch full-text search
- `GET /bible/v2/suggest?q={prefix}` - Predictive suggestions
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

**Examples:**

```bash
# Get verse in JSON format
curl "http://localhost:8000/bible/JOHN/3/16?json=true"

# Get verse with italics shown
curl "http://localhost:8000/bible/PSALM/23/1?show_italics=true"

# Combine parameters
curl "http://localhost:8000/bible/ROMANS/8/28?json=true&show_italics=true"
```

## 🛠️ Development

### Prerequisites

- Go 1.20+
- Redis server
- SQLite3
- OpenSearch 2.x (for v2 endpoints)

### Build from Source

```bash
# Install dependencies
go mod tidy

# Build
go build -o bible_api cmd/bible_api.go

# Create database (first time only)
./bible_api -createDB -dbPath ./data/kjv.db

# Run
./bible_api -dbPath ./data/kjv.db
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
REDIS_ADDR=localhost:6379    # Redis server address
REDIS_PASSWORD=              # Redis password (if any)
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
GOOGLE_CLIENT_ID=1087565480706-8ntgu6rrcbpfmtnlqd2pair903q664v5.apps.googleusercontent.com
```

## TODO:

- Swipe to next chapter
- Expand OpenSearch features (filters, relevance tuning)
- Detailed search analytics
- Authentication for admin endpoints
- Rate limiting per user (not just IP)
- WebSocket support for real-time updates
- Document OpenSearch DNS resolution errors (`getaddrinfo ENOTFOUND opensearch`) and how to fix by running all services via `docker compose` so they share the same network

## 📚 Legacy Text Sources

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

## 📁 Project Structure

```
bible_api/
├── cmd/
│   └── bible_api.go          # Main application entry point
├── kjv/
│   ├── kjv.go               # Core Bible API handlers
│   ├── admin.go             # Admin endpoints for rate limiting
│   └── templates.go         # HTML templates
├── middleware/
│   └── rate_limiter.go      # Redis-based rate limiting
├── assets/
│   └── texts/
│       ├── README.md
│       ├── asv/
│       ├── asvs/
│       ├── bishops/
│       ├── coverdale/
│       ├── geneva/
│       ├── kjv_strongs/
│       ├── net/
│       ├── tyndale/
│       └── web/
├── data/
│   └── kjv.db              # SQLite Bible database
├── docker-compose.yml       # Docker setup with Redis
├── scripts/
│   └── index_kjv_to_opensearch.py # Bulk index KJV DB into OpenSearch
├── start_with_redis.sh     # Development startup script
├── test_rate_limit.sh      # Rate limiting test script
├── test_logging.sh         # Logging test script
└── RATE_LIMITING.md        # Detailed rate limiting documentation
```

## 📖 Documentation

- [Rate Limiting Guide](RATE_LIMITING.md) - Detailed rate limiting documentation
- [API Examples](#-testing) - Example API calls and responses

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📜 License

This project is open source. The King James Version text is in the public domain.
