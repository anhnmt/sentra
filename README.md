# Sentra

> High-performance, extensible threat scanner for incident response and threat hunting

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-lightgrey?style=flat)]()

Sentra is a fast, single-binary threat scanner built for filesystem triage. It runs multiple detection engines in parallel, uses a split I/O and analysis worker pool to maximize throughput, and ships with auto-updating signatures. Designed to be extended — new detectors plug in alongside the existing ones without changing the core pipeline.

---

## Features

- **Multi-detector pipeline** — detectors run in parallel per file; results are aggregated and deduplicated across engines
- **YARA detection** — dual backend ([yara-x](https://github.com/VirusTotal/yara-x) + [yara](https://github.com/VirusTotal/yara)) with graceful fallback if one backend fails
- **High-performance architecture** — separate I/O pool and analysis pool (capped at `NumCPU`), mmap for large files, batch processing for small files, magic-byte pre-filter
- **Auto signature update** — fetches latest [YARA Forge Core](https://github.com/YARAHQ/yara-forge) rules with version caching to avoid redundant downloads
- **Smart filtering** — configurable file size range, skip directories, max walk depth, magic-byte filter for irrelevant binary formats
- **Graceful shutdown** — CTRL+C stops new work immediately, drains in-flight analysis safely before exit
- **Structured logging** — zerolog output with scan stats: files scanned, skipped, matches, errors, duration

---

## Installation

### From source

```bash
git clone https://github.com/anhnmt/sentra
cd sentra
go build -o sentra ./cmd/sentra
```

### Update signatures

```bash
./sentra --update-signatures
```

---

## Usage

```bash
# Scan a directory
./sentra --target /path/to/scan

# Scan with custom rules directory
./sentra --target /home --rules-dir /opt/rules

# Limit scan depth and skip build directories
./sentra --target /srv --max-depth 5 --skip-dir build --skip-dir dist

# Tune file size range (bytes)
./sentra --target /tmp --min-file-size 1024 --max-file-size 10485760

# Update signatures then scan
./sentra --update-signatures && ./sentra --target /var/www
```

---

## Options

### Scan options

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | *(required)* | File or directory to scan |
| `--rules-dir` | `signatures/yara` | YARA rules directory |
| `--workers` | `NumCPU × 2` | Number of I/O worker goroutines |
| `--skip-dir` | — | Directory name to skip, repeatable (`--skip-dir .git --skip-dir build`) |
| `--max-depth` | `0` *(unlimited)* | Maximum directory depth to walk |
| `--min-file-size` | `4096` (4 KB) | Minimum file size to scan in bytes |
| `--max-file-size` | `31457280` (30 MB) | Maximum file size to scan in bytes |

### Util

| Flag | Default | Description |
|------|---------|-------------|
| `--update-signatures` | `false` | Fetch latest YARA Forge Core rules |

---

## Architecture

```
Walk (fastwalk)
  └── eligibility filter (size, dir, magic bytes)
        └── ioPool  (Workers = NumCPU×2)
              ├── readFile: stat → ReadAll (<512KB) or mmap (≥512KB)
              ├── magic-byte check
              └── analysisPool  (Workers = NumCPU)
                    ├── detector A  ──┐
                    ├── detector B  ──┤
                    └── detector N  ──┴── dedup → result channel → consumer → log
```

**Two-pool design:** I/O is disk-bound and benefits from many goroutines. Analysis (especially CGo/native engines) is CPU-bound — capping it at `NumCPU` prevents thread thrashing and improves cache locality.

**mmap threshold (512 KB):** Small files use `io.ReadAll` to avoid mmap overhead. Large files use mmap to skip heap allocation and let the OS page cache do the work.

**Batch processing:** Files under the mmap threshold are grouped into batches of 16 before being submitted to the analysis pool, reducing goroutine spawn overhead for small-file-heavy directories.

**Detector interface:** Each detector implements a common interface — new detection engines plug into the same pipeline without touching the walker or worker logic.

---

## Detectors

| Detector | Status | Description |
|----------|--------|-------------|
| YARA (yara-x) | ✅ Active | Rust-based YARA engine, high performance |
| YARA (yarac) | ✅ Active | C-based YARA engine, broad rule compatibility |
| IOC matching | 🔜 Planned | Hash (MD5/SHA1/SHA256) and filename indicator matching |

---

## Skipped directories (built-in)

The following directories are always skipped in addition to any `--skip-dir` flags:

`.git` · `.svn` · `node_modules` · `vendor` · `.cache` · `.devenv`

---

## Signatures

Sentra uses [YARA Forge Core](https://github.com/YARAHQ/yara-forge) — a curated, high-accuracy ruleset aggregated from public repositories, optimized for low false positives.

Run `--update-signatures` to fetch the latest release. The current version is cached at `signatures/yara/.version` and compared against the GitHub API before downloading — no unnecessary re-downloads.

---

## Log output

```
INF scan starting target=/home workers=16 max_depth=0 min_file_size=4096 max_file_size=31457280
WRN match detected detector=yarax rule=Suspicious_PowerShell file=/home/user/malware.ps1
INF scan complete scanned=18432 skipped=3201 matches=1 errors=0 duration=12s
```

---

## Requirements

- Go 1.24+
- CGo enabled (required for yara-x and yara bindings)
- Linux, macOS, or Windows

---

## License

MIT