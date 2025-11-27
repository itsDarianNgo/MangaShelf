# Frequently Asked Questions

## General

### What is MangaShelf?

MangaShelf is a self-hosted application for downloading, organizing, and reading manga. It provides a web interface accessible from any device, automatic chapter updates, and a built-in reader.

### Is MangaShelf free?

Yes!  MangaShelf is free and open-source software, licensed under the MIT license. 

### What platforms does MangaShelf support?

MangaShelf runs on:
- Linux (x86_64, ARM64, ARMv7)
- macOS (Intel and Apple Silicon)
- Windows (x86_64)
- Docker (any platform)

### Does MangaShelf require an internet connection?

- **For downloading manga:** Yes
- **For reading downloaded manga:** No, the reader works offline

### Where does MangaShelf get manga from?

MangaShelf downloads from various online manga sources (MangaDex, MangaSee, etc.).  It does not host any content itself.

---

## Installation

### Do I need Docker?

No!  MangaShelf is a single binary with no external dependencies. Docker is optional and provided for convenience.

### Do I need a database server?

No.  MangaShelf uses an embedded SQLite database.  No PostgreSQL, MySQL, or other database server is required.

### Can I run MangaShelf on a Raspberry Pi?

Yes! MangaShelf provides ARM builds specifically for Raspberry Pi:
- **Pi 3/4 (64-bit OS):** Use `mangashelf-linux-arm64`
- **Pi 3/4 (32-bit OS):** Use `mangashelf-linux-armv7`
- **Pi Zero/1/2:** Use `mangashelf-linux-armv7`

See the [Raspberry Pi Guide](deployment/raspberry-pi.md) for details.

### Can I run MangaShelf on my NAS?

Yes!  MangaShelf works on:
- **Synology:** See [Synology Guide](deployment/synology.md)
- **QNAP:** Use Docker or the Linux binary
- **Unraid:** See [Unraid Guide](deployment/unraid.md)
- **TrueNAS:** Use Docker or jails

### How do I update MangaShelf?

**Binary installation:**
```bash
# Download the new version and replace the old binary
curl -sSL https://mangashelf.dev/install.sh | sh
```

**Docker:**
```bash
docker pull ghcr.io/username/mangashelf:latest
docker compose up -d
```

---

## Configuration

### Where is my data stored?

By default, data is stored in a `data` folder in the same directory as the binary:

```
./data/
├── mangashelf.db    # Database
├── manga/           # Downloaded manga
├── scrapers/        # Custom Lua scrapers
└── config.yaml      # Configuration (after first edit)
```

### How do I change the port?

```bash
# Command line
./mangashelf --port 9000

# Environment variable
MANGASHELF_SERVER_PORT=9000 ./mangashelf

# Config file
server:
  port: 9000
```

### How do I access MangaShelf from other devices?

By default, MangaShelf listens on all network interfaces (`0.0.0.0`). Access it using your computer's IP address:

```
http://YOUR_IP:8080
```

Make sure your firewall allows incoming connections on port 8080.

### How do I set up a reverse proxy? 

See the [Reverse Proxy Guide](reverse-proxy. md) for examples with:
- Nginx
- Caddy
- Traefik
- Apache

---

## Usage

### How do I add manga to my library? 

1. Click the **+ Add** button
2. Search for the manga name
3. Select from the results
4. Choose which chapters to download
5.  Click **Add to Library**

### How do I download new chapters?

**Automatic:** MangaShelf checks for new chapters based on your update schedule (default: every 6 hours). 

**Manual:** Click the refresh button on the manga detail page.

### How do I change the download format?

Go to **Settings → Downloads → Format** and choose:
- **CBZ** (default) - Comic book archive, works with most readers
- **PDF** - Portable document format
- **Raw** - Plain image files in folders

### Can I import my existing manga collection?

Yes! Place your manga folders in the library directory and MangaShelf will detect them:

```
<library>/
├── One Piece/
│   ├── Chapter 001.cbz
│   └── Chapter 002.cbz
└── Naruto/
    └── ... 
```

Then go to **Settings → Library → Scan Now**. 

### Can I import from Tachiyomi? 

Tachiyomi backup import is planned for a future release.  For now, you can:
1. Export your Tachiyomi library list
2.  Manually add the same manga in MangaShelf
3. Re-download the chapters

---

## Reader

### What reading modes are available?

- **Single Page:** One page at a time (default)
- **Double Page:** Two pages side by side (desktop)
- **Vertical Scroll:** Continuous scrolling (webtoon style)

### How do I change reading direction?

Click the settings icon in the reader and select:
- **Right to Left (RTL):** Traditional manga reading direction
- **Left to Right (LTR):** Western comic style

### Does the reader remember my progress?

Yes! MangaShelf automatically saves your reading position. When you return to a chapter, it will resume where you left off. 

### Can I read on my phone?

Yes! The web interface is fully responsive.  Access MangaShelf from your phone's browser using the same URL. 

For a more app-like experience, you can add MangaShelf to your home screen (PWA support coming soon).

---

## Sources

### Why can't I find a specific manga?

1. **Try different search terms** - Use the original Japanese title or alternative spellings
2. **Try different sources** - Not all sources have all manga
3. **Check if the manga exists** on the source website directly

### A source stopped working.  What do I do? 

1. **Check if the source website is up**
2. **Update MangaShelf** to the latest version
3. **Check GitHub issues** for known problems
4. **Report the issue** if it's not already known

### Can I add my own sources?

Yes! MangaShelf supports custom Lua scrapers. See the [Custom Scrapers Guide](custom-scrapers.md). 

### Why are some chapters missing? 

Possible reasons:
1.  **Language filter:** MangaShelf might be filtering to your preferred language
2. **Source doesn't have them:** Try a different source
3. **Chapters are premium/paid:** Some sources restrict certain chapters

---

## Troubleshooting

### MangaShelf won't start

**Check the logs:**
```bash
./mangashelf --verbose
```

**Common issues:**
- Port already in use: Try a different port with `--port 9000`
- Permission denied: Make the binary executable with `chmod +x mangashelf`
- Missing library folder: MangaShelf will create it automatically

### Downloads are failing

1. **Check your internet connection**
2. **The source might be rate limiting you:** Wait a few minutes
3. **The source might be down:** Try a different source
4.  **Clear the cache:** Settings → Advanced → Clear Cache

### Images aren't loading in the reader

1.  **The chapter might not be fully downloaded:** Check the download status
2. **The CBZ file might be corrupted:** Try re-downloading
3.  **Browser cache issue:** Hard refresh with Ctrl+Shift+R

### The web interface is slow

1. **Large library:** MangaShelf handles thousands of manga, but initial load might be slow
2.  **Many covers loading:** Covers are lazy-loaded, wait for them to cache
3. **Slow disk:** If using a HDD or network storage, consider using an SSD

### Database errors

If you encounter database errors:

1. **Stop MangaShelf**
2. **Backup your data folder**
3. **Run migrations:**
   ```bash
   ./mangashelf migrate
   ```
4. **If that fails, check GitHub issues or Discord**

---

## Security

### Is MangaShelf secure?

MangaShelf is designed for home/private network use. If exposing to the internet:

1. **Use a reverse proxy** with HTTPS
2. **Enable authentication** (see [Configuration](configuration.md))
3. **Keep MangaShelf updated**
4. **Use a firewall** to restrict access

### Does MangaShelf phone home?

No.  MangaShelf does not collect telemetry or send data anywhere.  The only network requests are to manga sources and (optionally) metadata providers like Anilist.

### Can multiple people use MangaShelf?

Currently, MangaShelf is single-user. Multi-user support with separate libraries and reading progress is planned for a future release. 

---

## Legal

### Is MangaShelf legal?

MangaShelf is a tool that downloads publicly available manga from online sources. The legality depends on:
- Your local laws regarding downloading copyrighted content
- The terms of service of the manga sources
- Whether the manga is legally free or licensed

**MangaShelf does not host any content. ** It is a personal tool similar to a web browser or download manager.

### I'm a content creator. How do I request removal? 

MangaShelf is just a download client. Please contact the source websites directly regarding content removal. 

---

## Still have questions?

- 📖 Check the [full documentation](README.md)
- 💬 Ask in [Discord](https://discord.gg/mangashelf)
- 🐛 Report issues on [GitHub](https://github.com/username/mangashelf/issues)