# yoink

A CLI tool for downloading comics from readallcomics.com and packaging them as `.cbz` archives.

## How it works

1. Fetches the comic page and extracts the title and image links
2. Downloads all pages concurrently with Cloudflare bypass
3. Packages the images into a `.cbz` (Comic Book Zip) archive
4. Cleans up downloaded images, keeping only the cover (`001`)

## Installation

Build from source (requires Go 1.22.3+):

```shell
go build -o yoink
```

Pre-built binaries for Linux (arm64) and Windows are available on the [releases page](../../releases).

## Usage

```shell
yoink <url>
```

**Example:**

```shell
yoink https://readallcomics.com/ultraman-x-avengers-001-2024/
```

The comic title is extracted from the page and used to name the archive. Output is saved to:

```
<library>/<Title>/<Title>.cbz
```

## Configuration

| Variable        | Default      | Description                          |
|-----------------|--------------|--------------------------------------|
| `YOINK_LIBRARY` | `~/.yoink`   | Directory where comics are stored    |

```shell
YOINK_LIBRARY=/mnt/media/comics yoink https://readallcomics.com/some-comic-001/
```

## Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery) — HTML parsing
- [cloudflare-bp-go](https://github.com/DaRealFreak/cloudflare-bp-go) — Cloudflare bypass
- [cobra](https://github.com/spf13/cobra) — CLI framework
