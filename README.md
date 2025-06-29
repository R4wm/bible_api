# Bible API

> A production-ready, high-performance RESTful API for the King James Version Bible

- **High Performance**: Built with Go for speed and reliability
- **King James Version**: Pure Cambridge Text
- **Clean & Fast**: No ads, no distractions, ever
- **Enterprise Ready**: Redis-based rate limiting and structured logging
- **Security First**: Input validation, SQL injection protection, and comprehensive error handling
- **Easy Navigation**: Intuitive REST endpoints with JSON/HTML responses
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
      - Yearly Reading schedule for Old Testament and New Testament for every day
      - Monthly Reading schedule for Proverbs and Psalms for every day

## ⚡ Features

### 📖 Bible Content

- Complete King James Version
- Book navigation with clickable chapters
- Verse range support (e.g., `/bible/romans/5/1-5`)
- Full-text search across all books
- Daily reading schedules for OT, NT, Psalms, and Proverbs

### 🛡️ Rate Limiting (NEW)

- **5 requests per second** per IP address
- **1-minute blocking** when limit exceeded
- Redis-based storage with automatic TTL cleanup
- Rate limit headers in all responses
- Admin endpoints for manual IP management
- Comprehensive logging of rate limit events

### 🔧 Technical Features

- **Dual Format Support**: JSON and HTML response formats
- **RESTful Design**: Clean, intuitive API endpoints
- **SQLite Backend**: Fast, reliable database with parameterized queries
- **Redis Integration**: Rate limiting and caching
- **Docker Ready**: Full docker-compose setup included
- **Production Grade**: Comprehensive error handling and input validation
- **Structured Logging**: JSON logging with configurable levels
- **Environment Config**: Flexible configuration via environment variables
- **Security Hardened**: SQL injection protection and input sanitization

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
- `GET /bible/random_verse` - Get random verse

### Daily Reading

- `GET /bible/daily/ot` - Old Testament daily reading
- `GET /bible/daily/nt` - New Testament daily reading
- `GET /bible/daily/psalms` - Daily Psalms
- `GET /bible/daily/proverbs` - Daily Proverbs

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

- Go 1.16+
- Redis server
- SQLite3

### Build from Source

```bash
# Install dependencies
go mod tidy

# Build
go build -o bible_api cmd/bible_api.go

# Create database (first time only)
./bible_api -createDB -dbPath ./data/kjv.db

# Run with default settings
./bible_api -dbPath ./data/kjv.db

# Run with custom port
./bible_api -dbPath ./data/kjv.db -port 9000

# Run with environment variables
export DB_PATH=./data/kjv.db
export PORT=8080
export LOG_LEVEL=debug
./bible_api
```

### Environment Variables

```bash
# Database Configuration
DB_PATH=/path/to/kjv.db      # Path to SQLite database file

# Server Configuration  
PORT=8000                    # Server port (default: 8000)

# Redis Configuration
REDIS_ADDR=localhost:6379    # Redis server address
REDIS_PASSWORD=              # Redis password (if any)

# Logging Configuration
LOG_LEVEL=info              # Log level: debug, info, warn, error
```

## 🗺️ Roadmap

### Upcoming Features
- [ ] Swipe navigation for mobile interfaces
- [ ] Elasticsearch integration for advanced search
- [ ] Detailed search analytics and metrics
- [ ] JWT authentication for admin endpoints
- [ ] User-based rate limiting (not just IP)
- [ ] WebSocket support for real-time updates
- [ ] API versioning support
- [ ] Prometheus metrics endpoint

### Recent Improvements ✅
- [x] **Security Hardening**: SQL injection prevention, input validation
- [x] **Enhanced Error Handling**: Graceful error responses, no more panics
- [x] **Structured Logging**: JSON logging with configurable levels
- [x] **Environment Configuration**: Flexible config via env vars
- [x] **Production Ready**: Comprehensive cleanup and optimization

## 📁 Project Structure

```
bible_api/
├── cmd/
│   └── bible_api.go          # Main application entry point with logging config
├── kjv/
│   ├── kjv.go               # Core Bible API handlers (security hardened)
│   ├── admin.go             # Admin endpoints for rate limiting
│   ├── reading.go           # Daily reading schedule logic
│   └── templates.go         # HTML templates
├── middleware/
│   └── rate_limiter.go      # Redis-based rate limiting with structured logging
├── data/
│   └── kjv.db              # SQLite Bible database
├── docker-compose.yml       # Docker setup with Redis
├── start_with_redis.sh     # Development startup script
├── test_rate_limit.sh      # Rate limiting test script
├── test_logging.sh         # Logging test script
└── RATE_LIMITING.md        # Detailed rate limiting documentation
```

### Key Architecture Features

- **Modular Design**: Clean separation of concerns
- **Security First**: Parameterized queries, input validation
- **Observability**: Structured logging throughout
- **Configuration**: Environment-based configuration
- **Error Handling**: Graceful error responses, no panics

## 📖 Documentation

- [Rate Limiting Guide](RATE_LIMITING.md) - Detailed rate limiting documentation
- [API Examples](#-testing) - Example API calls and responses
- [Security Features](#-technical-features) - Input validation and SQL injection protection
- [Configuration Guide](#environment-variables) - Environment variable reference

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📜 License

This project is open source. The King James Version text is in the public domain.
