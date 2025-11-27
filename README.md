<p align="center">
  <img src="docs/assets/logo.png" alt="MangaShelf Logo" width="128" height="128">
</p>

<h1 align="center">MangaShelf</h1>

<p align="center">
  <strong>A modern, self-hosted manga downloader and reader</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <img alt="GitHub release" src="https://img.shields.io/github/v/release/username/mangashelf? style=flat-square">
  <img alt="Go version" src="https://img.shields.io/github/go-mod/go-version/username/mangashelf? style=flat-square">
  <img alt="License" src="https://img.shields.io/github/license/username/mangashelf?style=flat-square">
  <img alt="Build status" src="https://img.shields.io/github/actions/workflow/status/username/mangashelf/build.yml?style=flat-square">
</p>

<p align="center">
  <img alt="Linux" src="https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black">
  <img alt="macOS" src="https://img. shields.io/badge/macOS-000000?style=for-the-badge&logo=apple&logoColor=white">
  <img alt="Windows" src="https://img.shields. io/badge/Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white">
  <img alt="Raspberry Pi" src="https://img.shields. io/badge/Raspberry%20Pi-A22846?style=for-the-badge&logo=raspberrypi&logoColor=white">
  <img alt="Docker" src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white">
</p>

<p align="center">
  <img src="docs/assets/screenshot-library.png" alt="MangaShelf Library View" width="800">
</p>

---

## Why MangaShelf?

| Problem | MangaShelf Solution |
|---------|---------------------|
| 🐳 Complex Docker setups with multiple containers | **Single binary** - just download and run |
| 📦 External dependencies (PostgreSQL, Redis) | **Self-contained** - embedded SQLite database |
| 💻 CLI-only tools intimidate casual users | **Beautiful web UI** - accessible from any device |
| 📖 Need separate apps to read downloaded manga | **Built-in reader** - download and read in one place |
| 🔧 Difficult configuration and setup | **Zero config** - works out of the box |

---

## Features

### 📚 Library Management
- Beautiful grid view of your manga collection
- Automatic cover art fetching
- Track reading progress across devices
- Smart organization by series, status, and tags

### 🔍 Multi-Source Search
- Search across multiple manga sources simultaneously
- Preview manga details before adding to library
- Language preference filtering
- Duplicate detection

### ⬇️ Powerful Downloader
- Concurrent downloads with configurable workers
- Automatic retry with exponential backoff
- Resume interrupted downloads
- Export to CBZ, PDF, or raw images

### 📖 Built-in Reader
- Multiple reading modes (single page, double spread, vertical scroll)
- Right-to-left and left-to-right support
- Keyboard shortcuts and touch gestures
- Auto-save reading position
- Night mode and customizable themes

### 🔄 Automatic Updates
- Scheduled chapter checking
- Per-manga update intervals
- New chapter notifications
- Background downloading

### 🔌 Extensible
- Built-in sources (MangaDex, MangaSee, Manganato)
- Lua scripting for custom sources
- REST API for integrations
- OPDS feed for e-reader apps

---

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://mangashelf.dev/install.sh | sh
```

### Download Binary

Download the latest release for your platform from the [Releases](https://github.com/username/mangashelf/releases) page:

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | x86_64 | [mangashelf-linux-amd64](https://github.com/username/mangashelf/releases/latest/download/mangashelf-linux-amd64) |
| Linux | ARM64 | [mangashelf-linux-arm64](https://github.com/username/mangashelf/releases/latest/download/mangashelf-linux-arm64) |
| Linux | ARMv7 | [mangashelf-linux-armv7](https://github. com/username/mangashelf/releases/latest/download/mangashelf-linux-armv7) |
| macOS | x86_64 | [mangashelf-darwin-amd64](https://github.com/username/mangashelf/releases/latest/download/mangashelf-darwin-amd64) |
| macOS | Apple Silicon | [mangashelf-darwin-arm64](https://github.com/username/mangashelf/releases/latest/download/mangashelf-darwin-arm64) |
| Windows | x86_64 | [mangashelf-windows-amd64.exe](https://github.com/username/mangashelf/releases/latest/download/mangashelf-windows-amd64.exe) |

### Docker

```bash
docker run -d \
  --name mangashelf \
  -p 8080:8080 \
  -v ./data:/data \
  ghcr.io/username/mangashelf:latest
```

Or with Docker Compose:

```yaml
# docker-compose.yml
version: '3'
services:
  mangashelf:
    image: ghcr.io/username/mangashelf:latest
    container_name: mangashelf
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - TZ=America/New_York
    restart: unless-stopped
```

```bash
docker compose up -d
```

### Package Managers

<details>
<summary><strong>Homebrew (macOS/Linux)</strong></summary>

```bash
brew install username/tap/mangashelf
```

</details>

<details>
<summary><strong>Arch Linux (AUR)</strong></summary>

```bash
yay -S mangashelf-bin
```

</details>

<details>
<summary><strong>Scoop (Windows)</strong></summary>

```powershell
scoop bucket add extras
scoop install mangashelf
```

</details>

<details>
<summary><strong>Nix</strong></summary>

```bash
nix-env -iA nixpkgs.mangashelf
```

</details>

### Build from Source

```bash
# Clone the repository
git clone https://github.com/username/mangashelf.git
cd mangashelf

# Build (requires Go 1.21+ and Node.js 18+)
make build

# Or install directly
make install
```

---

## Quick Start

### 1. Start the Server

```bash
# Using the binary
./mangashelf

# Or specify a custom data directory
./mangashelf --data /path/to/manga
```

### 2. Open the Web UI

Navigate to [http://localhost:8080](http://localhost:8080) in your browser.

### 3. Add Your First Manga

1.  Click the **+ Add** button in the top right
2. Search for a manga (e.g., "One Piece")
3. Select the manga and choose a source
4. Click **Add to Library**
5. Select chapters to download

### 4. Start Reading! 

Click on any downloaded chapter to open the built-in reader. 

<p align="center">
  <img src="docs/assets/screenshot-reader.png" alt="MangaShelf Reader" width="600">
</p>

---

## Configuration

MangaShelf works out of the box with sensible defaults. Configuration is optional. 

### Configuration File

On first run, a default configuration file is created at:
- **Linux/macOS**: `~/.config/mangashelf/config.yaml`
- **Windows**: `%APPDATA%\mangashelf\config.yaml`

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080

library:
  path: "./data/manga"

downloader:
  workers: 3
  format: "cbz"  # cbz, pdf, raw

updates:
  enabled: true
  interval: "0 */6 * * *"  # Every 6 hours

notifications:
  enabled: false
```

See [Configuration Reference](docs/configuration. md) for all options.

### Environment Variables

All settings can be overridden with environment variables:

```bash
export MANGASHELF_SERVER_PORT=9000
export MANGASHELF_LIBRARY_PATH=/mnt/manga
export MANGASHELF_DOWNLOADER_WORKERS=5
```

### Command Line Options

```bash
./mangashelf --help

Usage:
  mangashelf [command]

Commands:
  serve       Start the web server (default)
  version     Print version information
  migrate     Run database migrations
  export      Export library to JSON
  import      Import library from backup

Flags:
  -c, --config string   Path to config file
  -d, --data string     Path to data directory (default "./data")
  -p, --port int        Server port (default 8080)
  -v, --verbose         Enable verbose logging
  -h, --help            Help for mangashelf
```

---

## Documentation

📖 **[Full Documentation](docs/README.md)**

- [Getting Started Guide](docs/getting-started.md)
- [Configuration Reference](docs/configuration.md)
- [Supported Sources](docs/sources.md)
- [Custom Scrapers (Lua)](docs/custom-scrapers. md)
- [API Reference](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Troubleshooting](docs/troubleshooting.md)
- [FAQ](docs/faq.md)

---

## Screenshots

<details>
<summary><strong>📱 Mobile View</strong></summary>
<p align="center">
  <img src="docs/assets/screenshot-mobile.png" alt="Mobile View" width="300">
</p>
</details>

<details>
<summary><strong>📖 Reader Modes</strong></summary>
<p align="center">
  <img src="docs/assets/screenshot-reader-vertical.png" alt="Vertical Scroll Mode" width="600">
</p>
</details>

<details>
<summary><strong>🔍 Search</strong></summary>
<p align="center">
  <img src="docs/assets/screenshot-search. png" alt="Search" width="600">
</p>
</details>

<details>
<summary><strong>⚙️ Settings</strong></summary>
<p align="center">
  <img src="docs/assets/screenshot-settings.png" alt="Settings" width="600">
</p>
</details>

---

## Comparison

| Feature | MangaShelf | Kaizoku | Mangal | Tachiyomi |
|---------|:----------:|:-------:|:------:|:---------:|
| Web UI | ✅ | ✅ | ❌ | ❌ |
| Built-in Reader | ✅ | ❌ | ❌ | ✅ |
| Single Binary | ✅ | ❌ | ✅ | ❌ |
| No External DB | ✅ | ❌ | ✅ | ✅ |
| Self-Hosted | ✅ | ✅ | ✅ | ❌ |
| Desktop/Mobile | ✅ | ✅ | ❌ | 📱 only |
| Auto Updates | ✅ | ✅ | ❌ | ✅ |
| Notifications | ✅ | ✅ | ❌ | ✅ |
| Custom Sources | ✅ | ✅ | ✅ | ✅ |
| ARM Support | ✅ | ⚠️ | ✅ | ✅ |

---

## Integrations

### Media Servers

MangaShelf organizes downloads in a format compatible with popular media servers:

- **[Komga](https://komga.org/)** - Full compatibility with ComicInfo.xml metadata
- **[Kavita](https://www.kavitareader.com/)** - Automatic library detection
- **[Calibre](https://calibre-ebook. com/)** - Import via folder monitoring

### Tracking Services

- **[Anilist](https://anilist.co/)** - Sync reading progress, fetch metadata
- **[MyAnimeList](https://myanimelist. net/)** - (Coming soon)

### Notifications

MangaShelf supports 80+ notification services via [Apprise](https://github.com/caronc/apprise):

- Discord
- Telegram
- Slack
- Email
- Pushover
- And many more...

### E-Readers

Access your library from e-reader apps using the built-in OPDS feed:

```
http://your-server:8080/opds
```

Compatible apps: Librera, Moon+ Reader, Panels, Chunky, and more.

---

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/username/mangashelf.git
cd mangashelf

# Install dependencies
make deps

# Run in development mode (hot reload)
make dev

# Run tests
make test

# Build for all platforms
make build-all
```

### Project Structure

```
mangashelf/
├── cmd/mangashelf/     # Application entry point
├── internal/           # Private application code
│   ├── api/            # HTTP handlers and routing
│   ├── database/       # Database schema and queries
│   ├── downloader/     # Download engine
│   ├── library/        # Library management
│   ├── reader/         # Reader service
│   └── scraper/        # Source providers
├── web/                # Frontend (Svelte)
├── docs/               # Documentation
└── scripts/            # Build and utility scripts
```

---

## Roadmap

### v1.0 (Current Development)
- [x] Core library management
- [x] MangaDex, MangaSee, Manganato sources
- [x] Download queue with retry logic
- [x] Web-based reader
- [x] Basic settings UI
- [ ] Scheduled updates
- [ ] Notifications

### v1.1
- [ ] Anilist integration
- [ ] OPDS feed
- [ ] Lua custom scrapers
- [ ] Import from Tachiyomi backup

### v1.2
- [ ] Multi-user support
- [ ] PWA / Offline mode
- [ ] Reading statistics
- [ ] Collections and tags

### Future
- [ ] MyAnimeList integration
- [ ] Manga recommendations
- [ ] Social features (sharing, comments)
- [ ] Plugin system

See the [Project Board](https://github. com/username/mangashelf/projects/1) for detailed progress.

---

## Support

- 💬 **[Discord](https://discord.gg/mangashelf)** - Chat with the community
- 🐛 **[Issues](https://github. com/username/mangashelf/issues)** - Report bugs or request features
- 💡 **[Discussions](https://github. com/username/mangashelf/discussions)** - Ask questions and share ideas
- 📖 **[Wiki](https://github.com/username/mangashelf/wiki)** - Community documentation

---

## Acknowledgments

MangaShelf stands on the shoulders of giants:

- **[mangal](https://github. com/metafates/mangal)** - Inspiration for the scraper architecture
- **[Kaizoku](https://github.com/oae/kaizoku)** - Inspiration for the web UI approach
- **[Tachiyomi](https://github.com/tachiyomiorg/tachiyomi)** - Inspiration for the reading experience
- **[gopher-lua](https://github. com/yuin/gopher-lua)** - Lua VM for custom scrapers
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - (Reserved for potential TUI)

---

## License

MangaShelf is open-source software licensed under the [MIT License](LICENSE).

---

<p align="center">
  <sub>
    If you find MangaShelf useful, please consider giving it a ⭐ on GitHub!
  </sub>
</p>

<p align="center">
  <a href="https://star-history.com/#username/mangashelf&Date">
    <img src="https://api.star-history. com/svg?repos=username/mangashelf&type=Date" alt="Star History">
  </a>
</p>