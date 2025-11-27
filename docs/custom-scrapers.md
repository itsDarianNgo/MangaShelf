# Custom Scrapers

MangaShelf supports custom Lua scrapers, allowing you to add support for any manga source. This guide explains how to write, test, and install custom scrapers. 

## Overview

Scrapers are Lua scripts that define how to:
1. **Search** for manga on a source
2. **Fetch manga details** (title, cover, description)
3. **List chapters** for a manga
4.  **Get page URLs** for a chapter

## Quick Start

### 1. Create a Scraper File

Create a new file in your scrapers directory:

```bash
# Find your scrapers directory
./mangashelf config get sources.customPath
# Default: ./data/scrapers

# Create a new scraper
touch ./data/scrapers/my-source.lua
```

### 2. Basic Template

```lua
-- my-source.lua
-- Custom scraper for My Manga Source

local http = require("http")
local html = require("html")
local json = require("json")

-- Source information (required)
local info = {
    id = "my-source",           -- Unique identifier
    name = "My Manga Source",   -- Display name
    baseUrl = "https://example-manga.com",
    languages = {"en"},         -- Supported languages
    nsfw = false                -- Contains adult content? 
}

-- Search for manga
-- @param query string - Search term
-- @return table - Array of manga results
local function search(query)
    local url = info.baseUrl .. "/search?q=" ..  http.encode(query)
    local response = http. get(url)
    
    if response.status ~= 200 then
        error("Search failed: " .. response.status)
    end
    
    local doc = html.parse(response.body)
    local results = {}
    
    for _, element in ipairs(doc:select(".manga-card")) do
        table.insert(results, {
            id = element:attr("data-id"),
            title = element:select(".title"):text(),
            cover = element:select("img"):attr("src"),
            url = element:select("a"):attr("href")
        })
    end
    
    return results
end

-- Get manga details
-- @param id string - Manga ID
-- @return table - Manga details
local function getManga(id)
    local url = info.baseUrl .. "/manga/" .. id
    local response = http. get(url)
    local doc = html.parse(response.body)
    
    return {
        id = id,
        title = doc:select("h1. title"):text(),
        description = doc:select(". description"):text(),
        cover = doc:select(". cover img"):attr("src"),
        status = doc:select(". status"):text(),
        author = doc:select(".author"):text(),
        artist = doc:select(". artist"):text(),
        genres = parseGenres(doc:select(".genres")),
        url = url
    }
end

-- Get chapter list
-- @param mangaId string - Manga ID
-- @return table - Array of chapters
local function getChapters(mangaId)
    local url = info.baseUrl .. "/manga/" .. mangaId ..  "/chapters"
    local response = http. get(url)
    local doc = html.parse(response. body)
    
    local chapters = {}
    
    for i, element in ipairs(doc:select(". chapter-item")) do
        table.insert(chapters, {
            id = element:attr("data-id"),
            title = element:select(".chapter-title"):text(),
            number = tonumber(element:attr("data-number")) or i,
            volume = element:attr("data-volume") or "",
            url = element:select("a"):attr("href"),
            publishedAt = element:attr("data-date")
        })
    end
    
    return chapters
end

-- Get page URLs for a chapter
-- @param chapterId string - Chapter ID
-- @return table - Array of pages
local function getPages(chapterId)
    local url = info.baseUrl ..  "/chapter/" .. chapterId
    local response = http.get(url)
    local doc = html.parse(response.body)
    
    local pages = {}
    
    for i, element in ipairs(doc:select(".page-image")) do
        table.insert(pages, {
            index = i,
            url = element:attr("src"),
            filename = string.format("%03d. jpg", i)
        })
    end
    
    return pages
end

-- Helper function to parse genres
local function parseGenres(element)
    local genres = {}
    for _, genre in ipairs(element:select("a")) do
        table. insert(genres, genre:text())
    end
    return genres
end

-- Export the scraper
return {
    info = info,
    search = search,
    getManga = getManga,
    getChapters = getChapters,
    getPages = getPages
}
```

### 3.  Test Your Scraper

```bash
# Test search
./mangashelf scraper test my-source search "one piece"

# Test manga details
./mangashelf scraper test my-source manga <manga-id>

# Test chapter list
./mangashelf scraper test my-source chapters <manga-id>

# Test pages
./mangashelf scraper test my-source pages <chapter-id>
```

## Available Libraries

MangaShelf provides these Lua libraries for scrapers:

### http

HTTP client for making web requests.

```lua
local http = require("http")

-- GET request
local response = http.get("https://example.com")
print(response.status)  -- 200
print(response.body)    -- HTML content

-- GET with headers
local response = http.get("https://example.com", {
    headers = {
        ["User-Agent"] = "MangaShelf/1.0",
        ["Accept"] = "application/json"
    }
})

-- POST request
local response = http.post("https://example.com/api", {
    headers = {
        ["Content-Type"] = "application/json"
    },
    body = json.encode({query = "one piece"})
})

-- URL encoding
local encoded = http.encode("search term with spaces")
-- Result: "search%20term%20with%20spaces"
```

### html

HTML parser for extracting data from web pages.

```lua
local html = require("html")

-- Parse HTML
local doc = html.parse("<html><body><h1>Hello</h1></body></html>")

-- Select elements (CSS selectors)
local heading = doc:select("h1")
print(heading:text())  -- "Hello"

-- Select multiple elements
for _, element in ipairs(doc:select(".item")) do
    print(element:text())
end

-- Get attributes
local link = doc:select("a")
print(link:attr("href"))

-- Check if element exists
if doc:select(".optional"):exists() then
    -- Element found
end

-- Get inner HTML
local content = doc:select(". 