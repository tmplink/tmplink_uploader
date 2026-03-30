# TmpLink Uploader

A simple and powerful file uploader for [TmpLink](https://tmp.link/), supporting large files, resumable uploads, and batch uploading.

**English** | [简体中文](README.zh-CN.md) | [日本語](README.ja.md)

## Overview

### Upload files via CLI

![CLI uploader](docs/images/tmplink-cli.webp)

### Upload files via TUI

![TUI uploader](docs/images/tmplink.webp)

## Features

- **One-click upload** - Supports files up to 50GB
- **High speed** - Chunked upload + multi-threading for maximum bandwidth utilization
- **Resumable uploads** - Automatically recovers from network interruptions
- **Dual interface** - TUI for interactive use, CLI for scripting and automation
- **Member features** - Sponsors enjoy additional advanced settings

## Quick Start

### Installation

#### One-click install (recommended)

**Option 1: Online install script**

Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/install-linux.sh | bash
```

macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/install-macos.sh | bash
```

Windows (PowerShell as Administrator):

```powershell
iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/install-windows.ps1'))
```

**Option 2: Clone and install**

```bash
git clone https://github.com/tmplink/tmplink_uploader.git
cd tmplink_uploader
```

Linux:

```bash
./install-linux.sh
```

macOS:

```bash
./install-macos.sh
```

Windows:

```powershell
.\install-windows.ps1
```

After installation, `tmplink` and `tmplink-cli` will be available system-wide.

#### Manual download

Pre-built binaries are available in the [build](build/) directory:

- **Windows**: `windows-64bit` or `windows-32bit`
- **macOS**: `macos-arm64` (M1/M2) or `macos-intel`
- **Linux**: `linux-64bit`, `linux-32bit`, or `linux-arm64`

### Get your access token

1. Open [TmpLink](https://tmp.link/) and log in
2. Click "Upload File", then click "Reset"
3. Copy your Token from the "CLI Upload" section

### Usage

#### TUI (recommended for beginners)

```bash
# If installed via script:
tmplink

# If manually downloaded:
# Windows
tmplink.exe

# macOS/Linux
./tmplink
```

On first launch you will be prompted for your Token. After that, use the interface to select and upload files.

#### CLI (for advanced users)

```bash
# Save your Token (one-time setup)
tmplink-cli -set-token YOUR_TOKEN

# Upload a file
tmplink-cli -file /path/to/file
```

## Examples

### TUI navigation

Use arrow keys to navigate after launch:

- **Select file** — Browse and choose the file to upload
- **Start upload** — Monitor real-time progress and speed
- **View result** — Copy the download link

### Common CLI usage

```bash
# Upload a single file
tmplink-cli -file ~/Documents/report.pdf

# Upload a large file with bigger chunks
tmplink-cli -file ~/Videos/movie.mp4 -chunk-size 10

# Keep a file permanently
tmplink-cli -file ~/backup.zip -model 99

# Use a temporary token for a single upload
tmplink-cli -file test.txt -token TEMP_TOKEN
```

## Parameters

### Basic parameters

| Flag | Description |
| --- | --- |
| `-file` | Path to the file (required) |
| `-token` | API Token (can be saved to config) |
| `-chunk-size` | Chunk size in MB, 1–99 (default: 3) |
| `-model` | Retention period: 0=24h, 1=3 days, 2=7 days, 99=permanent |

### Configuration

| Flag | Description |
| --- | --- |
| `-set-token` | Save Token to config file |
| `-set-model` | Set default retention period |
| `-set-mr-id` | Set default upload directory |

## Troubleshooting

### macOS: "Cannot verify developer"

```bash
xattr -d com.apple.quarantine tmplink tmplink-cli
```

### Linux/macOS: Permission denied

```bash
chmod +x tmplink tmplink-cli
```

### Windows Defender warning

Click "More info" → "Run anyway", or add the binary to your trusted list.

### Upload failures

1. Check your network connection
2. Verify your Token is valid
3. Use `-debug` for detailed error output

```bash
tmplink-cli -debug -file test.txt
```

## Getting help

```bash
# List all available flags
tmplink-cli -h

# Enable verbose logging
tmplink-cli -debug -file yourfile.txt
```

Still stuck? [Open an issue](https://github.com/tmplink/tmplink_uploader/issues)

## License

This project is licensed under [Apache 2.0](LICENSE).
