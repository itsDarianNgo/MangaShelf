# Configuration Reference

MangaShelf is designed to work out of the box with sensible defaults. This page documents all available configuration options for when you need to customize behavior.

## Configuration Methods

MangaShelf can be configured in three ways (in order of precedence):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (lowest priority)

## Configuration File

### Location

The configuration file is automatically created on first run at:

| Platform | Location |
|----------|----------|
| Linux | `~/.config/mangashelf/config.yaml` |
| macOS | `~/Library/Application Support/mangashelf/config.yaml` |
| Windows | `%APPDATA%\mangashelf\config.yaml` |
| Docker | `/data/config.yaml` |

You can specify a custom location:

```bash
./mangashelf --config /path/to/config.yaml
```

### Full Configuration Example

```yaml
# MangaShelf Configuration
# All values shown are defaults

#───────────────────────────────────────────────────────────────
# Server Configuration
#───────────────────────────────────────────────────────────────
server:
  # Host address to bind to
  # Use "0.0.0.0" to listen on all interfaces
  # Use "127.0.0.1" to only allow local connections
  host: "0.0.0.0"
  
  # Port to listen on
  port: 8080
  
  # Base URL for the application (used for generating links)
  # Set this if running behind a reverse proxy with a subpath
  # Example: "/mangashelf" if accessed at https://example.com/mangashelf
  baseUrl: ""
  
  # Enable CORS for API requests
  # Set to your frontend URL in development
  cors:
    enabled: false
    origins: []

#───────────────────────────────────────────────────────────────
# Library Configuration  
#───────────────────────────────────────────────────────────────
library:
  # Path where manga will be downloaded and stored
  path: "./data/manga"
  
  # Scan library for changes on startup
  scanOnStartup: true
  
  # Watch library folder for external changes
  watchForChanges: false

#───────────────────────────────────────────────────────────────
# Downloader Configuration
#───────────────────────────────────────────────────────────────
downloader:
  # Number of concurrent download workers
  # Increase for faster downloads, decrease to reduce server load
  workers: 3
  
  # Maximum retry attempts for failed downloads
  retryAttempts: 3
  
  # Delay between retry attempts
  retryDelay: "5s"
  
  # Request timeout for downloading pages
  timeout: "30s"
  
  # Rate limit (requests per second per source)
  # Helps avoid being blocked by sources
  rateLimit: "2/s"
  
  # User agent string for requests
  userAgent: "MangaShelf/1.0"

#───────────────────────────────────────────────────────────────
# Format Configuration
#───────────────────────────────────────────────────────────────
formats:
  # Default export format: "cbz", "pdf", or "raw"
  default: "cbz"
  
  # Compress images to reduce file size
  compressImages: false
  
  # JPEG quality when compressing (1-100)
  jpegQuality: 85
  
  # Maximum image width (0 = no resize)
  # Useful for saving space on mobile-focused libraries
  maxImageWidth: 0
  
  # Generate ComicInfo.xml metadata in CBZ files
  generateComicInfo: true

#───────────────────────────────────────────────────────────────
# Update Configuration
#───────────────────────────────────────────────────────────────
updates:
  # Enable automatic chapter checking
  enabled: true
  
  # Default update interval (cron expression)
  # "0 */6 * * *" = every 6 hours
  # "0 0 * * *" = daily at midnight
  # "0 0 * * 0" = weekly on Sunday
  defaultInterval: "0 */6 * * *"
  
  # Check for updates on startup
  checkOnStartup: true
  
  # Automatically download new chapters
  autoDownload: true

#───────────────────────────────────────────────────────────────
# Metadata Configuration
#───────────────────────────────────────────────────────────────
metadata:
  # Fetch metadata from Anilist
  fetchAnilist: true
  
  # Anilist API settings (optional, for higher rate limits)
  anilist:
    clientId: ""
    clientSecret: ""
  
  # Download cover images
  downloadCovers: true
  
  # Preferred cover size: "medium", "large", "extraLarge"
  coverSize: "large"

#───────────────────────────────────────────────────────────────
# Notification Configuration
#───────────────────────────────────────────────────────────────
notifications:
  # Enable notifications
  enabled: false
  
  # Apprise notification URL
  # See: https://github. com/caronc/apprise#supported-notifications
  # Examples:
  #   Discord: "discord://webhook_id/webhook_token"
  #   Telegram: "tgram://bot_token/chat_id"
  #   Email: "mailto://user:pass@gmail.com"
  apprise:
    urls: []
  
  # Events to notify on
  events:
    newChapters: true
    downloadComplete: false
    downloadFailed: true

#───────────────────────────────────────────────────────────────
# Reader Configuration
#───────────────────────────────────────────────────────────────
reader:
  # Default reading mode: "single", "double", "vertical"
  defaultMode: "single"
  
  # Default reading direction: "rtl" (right-to-left), "ltr" (left-to-right)
  defaultDirection: "rtl"
  
  # Preload adjacent pages for smoother reading
  preloadPages: 2
  
  # Remember reading position per chapter
  saveProgress: true

#───────────────────────────────────────────────────────────────
# Sources Configuration
#───────────────────────────────────────────────────────────────
sources:
  # Path to custom Lua scrapers
  customPath: "./data/scrapers"
  
  # Default source for searching
  default: "mangadex"
  
  # Source-specific settings
  mangadex:
    # Preferred language (ISO 639-1 code)
    language: "en"
    # Include NSFW content
    nsfw: false
    # Show chapters without available scanlation
    showUnavailable: false

#───────────────────────────────────────────────────────────────
# Logging Configuration
#───────────────────────────────────────────────────────────────
logging:
  # Log level: "debug", "info", "warn", "error"
  level: "info"
  
  # Log format: "text", "json"
  format: "text"
  
  # Write logs to file
  file:
    enabled: false
    path: "./data/logs/mangashelf.log"
    # Maximum log file size before rotation
    maxSize: "100MB"
    # Number of old log files to keep
    maxBackups: 3

#───────────────────────────────────────────────────────────────
# Database Configuration
#───────────────────────────────────────────────────────────────
database:
  # SQLite database path
  path: "./data/mangashelf.db"
  
  # Enable WAL mode for better concurrent performance
  walMode: true

#───────────────────────────────────────────────────────────────
# Security Configuration (Optional)
#───────────────────────────────────────────────────────────────
# security:
#   # Enable basic authentication
#   auth:
#     enabled: false
#     username: "admin"
#     password: "changeme"  # Use environment variable in production! 
#   
#   # Enable API key authentication
#   apiKey:
#     enabled: false
#     key: ""  # Generate with: openssl rand -hex 32
```

## Environment Variables

Every configuration option can be set via environment variable using the pattern:

```
MANGASHELF_<SECTION>_<KEY>=value
```

### Common Environment Variables

```bash
# Server
MANGASHELF_SERVER_HOST=0. 0.0.0
MANGASHELF_SERVER_PORT=8080
MANGASHELF_SERVER_BASEURL=/manga

# Library
MANGASHELF_LIBRARY_PATH=/data/manga

# Downloader
MANGASHELF_DOWNLOADER_WORKERS=5
MANGASHELF_DOWNLOADER_RATELIMIT=5/s

# Formats
MANGASHELF_FORMATS_DEFAULT=cbz
MANGASHELF_FORMATS_COMPRESSIMAGES=true

# Updates
MANGASHELF_UPDATES_ENABLED=true
MANGASHELF_UPDATES_DEFAULTINTERVAL="0 */4 * * *"

# Notifications
MANGASHELF_NOTIFICATIONS_ENABLED=true
MANGASHELF_NOTIFICATIONS_APPRISE_URLS="discord://xxx,tgram://yyy"

# Logging
MANGASHELF_LOGGING_LEVEL=debug
```

### Docker Environment Example

```yaml
# docker-compose.yml
services:
  mangashelf:
    image: ghcr. io/username/mangashelf:latest
    environment:
      - MANGASHELF_SERVER_PORT=8080
      - MANGASHELF_LIBRARY_PATH=/data/manga
      - MANGASHELF_DOWNLOADER_WORKERS=5
      - MANGASHELF_UPDATES_DEFAULTINTERVAL=0 */2 * * *
      - MANGASHELF_NOTIFICATIONS_ENABLED=true
      - MANGASHELF_NOTIFICATIONS_APPRISE_URLS=discord://webhook_id/webhook_token
      - TZ=America/New_York
    volumes:
      - ./data:/data
    ports:
      - "8080:8080"
```

## Command-Line Flags

```bash
./mangashelf --help

Flags:
  -c, --config string       Path to config file
  -d, --data string         Data directory (default "./data")
  -H, --host string         Server host (default "0.0.0. 0")
  -p, --port int            Server port (default 8080)
  -w, --workers int         Download workers (default 3)
  -v, --verbose             Enable debug logging
      --no-update           Disable automatic updates
      --no-scan             Skip library scan on startup
  -h, --help                Show help
```

### Examples

```bash
# Run on a different port
./mangashelf --port 9000

# Use a specific data directory
./mangashelf --data /mnt/storage/manga

# Enable debug logging
./mangashelf --verbose

# Combine multiple flags
./mangashelf --port 9000 --data /mnt/manga --workers 5 --verbose
```

## Cron Expression Reference

Update intervals use cron expressions:

| Expression | Description |
|------------|-------------|
| `0 * * * *` | Every hour |
| `0 */2 * * *` | Every 2 hours |
| `0 */6 * * *` | Every 6 hours |
| `0 0 * * *` | Daily at midnight |
| `0 0 * * 0` | Weekly on Sunday |
| `0 0 1 * *` | Monthly on the 1st |
| `never` | Disable automatic updates |

### Cron Format

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday = 0)
│ │ │ │ │
* * * * *
```

## File Organization

MangaShelf organizes downloaded manga in this structure:

```
<library. path>/
├── One Piece/
│   ├── cover.jpg
│   ├── series.json
│   ├── Chapter 0001.cbz
│   ├── Chapter 0002.cbz
│   └── ... 
├── Chainsaw Man/
│   ├── cover.jpg
│   ├── series. json
│   └── ... 
└── . mangashelf/
    ├── cache/
    └── temp/
```

## Next Steps

- [Reverse Proxy Setup](reverse-proxy.md) - Configure Nginx, Caddy, or Traefik
- [Notifications](notifications.md) - Set up alerts for new chapters
- [Sources](sources.md) - Configure manga sources