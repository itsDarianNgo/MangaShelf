# Supported Sources

MangaShelf supports multiple manga sources out of the box, with the ability to add custom sources via Lua scripts. 

## Built-in Sources

### MangaDex

**Website:** [mangadex.org](https://mangadex.org)

The largest and most popular scanlation platform. High quality scans with excellent metadata. 

| Feature | Support |
|---------|---------|
| Search | ✅ |
| Multiple languages | ✅ (40+ languages) |
| Metadata | ✅ (Excellent) |
| Rate limiting | Generous |
| NSFW | Optional |

**Configuration:**

```yaml
sources:
  mangadex:
    language: "en"      # ISO 639-1 language code
    nsfw: false         # Include adult content
    showUnavailable: false
```

**Supported Languages:**

`en`, `ja`, `ko`, `zh`, `zh-hk`, `es`, `es-la`, `fr`, `de`, `it`, `pt`, `pt-br`, `ru`, `pl`, `th`, `vi`, `id`, `tr`, `ar`, `hi`, and many more. 

---

### MangaSee

**Website:** [mangasee123.com](https://mangasee123. com)

Fast, reliable source with good coverage of popular titles.

| Feature | Support |
|---------|---------|
| Search | ✅ |
| Multiple languages | ❌ (English only) |
| Metadata | ⚠️ (Basic) |
| Rate limiting | Moderate |
| NSFW | Some |

---

### Manganato

**Website:** [manganato.com](https://manganato.com)

Large library with frequent updates.  Good for finding obscure titles.

| Feature | Support |
|---------|---------|
| Search | ✅ |
| Multiple languages | ❌ (English only) |
| Metadata | ⚠️ (Basic) |
| Rate limiting | Strict |
| NSFW | Some |

---

### MangaPill

**Website:** [mangapill.com](https://mangapill.com)

Alternative source with good selection. 

| Feature | Support |
|---------|---------|
| Search | ✅ |
| Multiple languages | ❌ (English only) |
| Metadata | ⚠️ (Basic) |
| Rate limiting | Moderate |
| NSFW | Some |

---

## Source Selection

When adding manga to your library, you can choose which source to use:

1. **Search results show source badges** - Easily identify where each result comes from
2. **Multiple sources for same title** - Choose based on image quality, translation, or update speed
3. **Change source later** - Switch sources from the manga settings page

### Choosing the Right Source

| Priority | Choose |
|----------|--------|
| Best quality | MangaDex |
| Fastest updates | Varies by title |
| Rare/obscure titles | Manganato, MangaSee |
| Non-English languages | MangaDex |

## Custom Sources (Lua Scrapers)

You can add support for additional sources by writing Lua scrapers. 

### Installing Community Scrapers

```bash
# Browse available scrapers
./mangashelf sources list --remote

# Install a scraper
./mangashelf sources install <scraper-name>
```

### Scraper Location

Custom scrapers are stored in:

```
<data-dir>/scrapers/
├── source1. lua
├── source2.lua
└── ...
```

### Writing a Custom Scraper

See [Custom Scrapers Guide](custom-scrapers.md) for detailed instructions.

**Basic template:**

```lua
-- my-source.lua
local http = require("http")
local html = require("html")

return {
    -- Source metadata
    info = {
        id = "my-source",
        name = "My Manga Source",
        baseUrl = "https://example.com",
        languages = {"en"},
        nsfw = false
    },

    -- Search for manga
    search = function(query)
        local response = http.get(info.baseUrl .. "/search? q=" .. query)
        local doc = html.parse(response.body)
        
        local results = {}
        for _, el in ipairs(doc:select(". manga-item")) do
            table.insert(results, {
                id = el:attr("data-id"),
                title = el:select(". title"):text(),
                cover = el:select("img"):attr("src"),
                url = el:select("a"):attr("href")
            })
        end
        return results
    end,

    -- Get manga details
    getManga = function(id)
        -- Implementation
    end,

    -- Get chapter list
    getChapters = function(mangaId)
        -- Implementation
    end,

    -- Get page URLs
    getPages = function(chapterId)
        -- Implementation
    end
}
```

## Troubleshooting Sources

### Source not returning results

1. **Check your internet connection**
2.  **The source website might be down** - Try accessing it directly
3. **You might be rate limited** - Wait a few minutes
4. **Clear the source cache:**
   ```bash
   ./mangashelf sources clear-cache <source-name>
   ```

### Images not loading

1. **Source might require specific headers** - Check scraper configuration
2. **Images might be hotlink protected** - Use the built-in proxy
3. **Source changed their structure** - Update the scraper

### Source is slow

1. **Reduce download workers** to avoid overwhelming the source
2. **Increase rate limit delay** in configuration
3. **Try a different source** for the same manga

### Source stopped working

Websites change frequently.  Check:

1.  **GitHub issues** for reported problems
2. **Update MangaShelf** to the latest version
3. **Update custom scrapers** if using any
4. **Report the issue** if it's a built-in source

## Source Status

Check the current status of all sources:

```bash
./mangashelf sources status
```

Output:
```
Source          Status      Last Check
─────────────────────────────────────────
MangaDex        ✅ Online   2 minutes ago
MangaSee        ✅ Online   2 minutes ago
Manganato       ⚠️ Slow     2 minutes ago
MangaPill       ✅ Online   2 minutes ago
```

## Next Steps

- [Custom Scrapers](custom-scrapers. md) - Write your own source
- [Source Troubleshooting](source-troubleshooting.md) - Fix source issues
- [Configuration](configuration.md) - Configure source settings