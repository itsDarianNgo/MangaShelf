# Getting Started with MangaShelf

This guide will walk you through installing MangaShelf and adding your first manga in under 5 minutes. 

## Prerequisites

- A computer running Linux, macOS, or Windows
- A web browser
- ~50MB of disk space for the application
- Additional storage for your manga library

## Step 1: Download MangaShelf

### Option A: Quick Install Script (Linux/macOS)

Open a terminal and run:

```bash
curl -sSL https://mangashelf. dev/install.sh | sh
```

This will:
- Detect your operating system and architecture
- Download the appropriate binary
- Install it to `/usr/local/bin/mangashelf`
- Make it executable

### Option B: Manual Download

1. Go to the [Releases page](https://github. com/username/mangashelf/releases/latest)
2. Download the binary for your platform:
   - **Linux (x86_64)**: `mangashelf-linux-amd64`
   - **Linux (ARM64/Raspberry Pi 4)**: `mangashelf-linux-arm64`
   - **macOS (Intel)**: `mangashelf-darwin-amd64`
   - **macOS (Apple Silicon)**: `mangashelf-darwin-arm64`
   - **Windows**: `mangashelf-windows-amd64.exe`

3. Make it executable (Linux/macOS):
   ```bash
   chmod +x mangashelf-linux-amd64
   ```

### Option C: Docker

```bash
docker run -d \
  --name mangashelf \
  -p 8080:8080 \
  -v ~/manga:/data \
  ghcr.io/username/mangashelf:latest
```

## Step 2: Start MangaShelf

### Linux/macOS

```bash
./mangashelf
```

### Windows

Double-click `mangashelf-windows-amd64. exe` or run from Command Prompt:

```cmd
mangashelf-windows-amd64.exe
```

### What You'll See

```
   __  ___                       _____ __         __  ____
  /  |/  /___ _____  ____ _____ / ___// /_  ___  / / / __/
 / /|_/ / __ `/ __ \/ __ `/ __ \\__ \/ __ \/ _ \/ / / /_  
/ /  / / /_/ / / / / /_/ / /_/ /__/ / / / /  __/ / / __/  
/_/  /_/\__,_/_/ /_/\__, /\__,_/____/_/ /_/\___/_/ /_/    
                   /____/                                  

→ Server starting on http://0.0.0. 0:8080
→ Data directory: ./data
→ Database initialized
→ Ready to serve manga!  📚
```

## Step 3: Open the Web Interface

Open your web browser and navigate to:

```
http://localhost:8080
```

You should see the MangaShelf welcome screen:

![Welcome Screen](assets/screenshot-welcome. png)

## Step 4: Add Your First Manga

1. **Click the "+ Add Manga" button** in the top right corner

2. **Search for a manga** by typing in the search box (e.g., "One Piece")

3. **Browse results** from different sources:
   
   ![Search Results](assets/screenshot-search-results.png)

4. **Click on a result** to see manga details:
   - Cover image
   - Synopsis
   - Author/Artist
   - Chapter count
   - Genres

5. **Click "Add to Library"** to add the manga

6. **Select chapters to download**:
   - Click individual chapters, or
   - Click "Select All" for everything, or
   - Use the range selector for specific chapters

7.  **Click "Download Selected"** to start downloading

## Step 5: Read Your Manga

Once chapters are downloaded:

1. Click on the manga in your library
2. Click on a chapter to open the reader
3. Use arrow keys or swipe to navigate pages

### Reader Controls

| Action | Keyboard | Mouse/Touch |
|--------|----------|-------------|
| Next page | `→` or `Space` | Click right side / Swipe left |
| Previous page | `←` | Click left side / Swipe right |
| Next chapter | `]` | Click "Next Chapter" |
| Previous chapter | `[` | Click "Previous Chapter" |
| Toggle fullscreen | `F` | Click fullscreen icon |
| Show/hide UI | `Esc` | Tap center |
| Close reader | `Q` | Click X |

## Next Steps

Now that you have MangaShelf running:

- **[Configure automatic updates](updates.md)** to check for new chapters
- **[Set up notifications](notifications.md)** to get alerts for new releases
- **[Customize settings](configuration.md)** to match your preferences
- **[Add more sources](sources. md)** for greater manga selection

## Troubleshooting

### Port already in use

If port 8080 is in use, specify a different port:

```bash
./mangashelf --port 9000
```

### Permission denied

Make sure the binary is executable:

```bash
chmod +x mangashelf
```

### Can't access from other devices

By default, MangaShelf binds to all interfaces. Make sure:
1. Your firewall allows port 8080
2. You're using the correct IP address of the host machine

```bash
# Find your IP address
ip addr show  # Linux
ipconfig      # Windows
```

Then access via `http://YOUR_IP:8080`

### Need more help?

- Check the [Troubleshooting Guide](troubleshooting.md)
- See the [FAQ](faq.md)
- Join our [Discord](https://discord.gg/mangashelf)