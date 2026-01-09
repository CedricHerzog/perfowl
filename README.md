# PerfOwl

**See what others miss in your performance profiles.**

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Welcome

PerfOwl is a powerful CLI tool and MCP server for analyzing browser profiler exports. Whether you're debugging performance issues, optimizing web applications, or integrating performance analysis into your AI workflows, PerfOwl helps you identify bottlenecks and understand complex traces.

## What is PerfOwl?

PerfOwl (**Optimization Workbench & Lab**) is a performance analysis toolkit that provides:

- **Bottleneck Detection** - GC pressure, layout thrashing, sync IPC, long tasks, network blocking
- **Extension Analysis** - Per-extension performance impact, DOM events, IPC messages
- **Call Tree Analysis** - Hot functions by self time and running time, hot path detection
- **Category Breakdown** - Time spent per profiler category (JavaScript, Layout, GC/CC, Network, etc.)
- **Thread Analysis** - CPU time, sample counts, wake patterns, category distribution
- **Worker Analysis** - Web Worker performance, CPU/idle time, messaging patterns, sync points
- **Crypto Profiling** - SubtleCrypto API usage, algorithm detection, serialization issues
- **Contention Detection** - GC pauses affecting workers, sync IPC blocking, lock contention
- **Scaling Analysis** - Parallel efficiency, speedup measurement, bottleneck identification
- **Profile Comparison** - Compare two profiles to identify improvements or regressions
- **MCP Server** - Integration with Claude and other AI assistants

Currently supports Firefox Profiler exports, with more browsers coming soon.

## Table of Contents

- [PerfOwl](#perfowl)
  - [Welcome](#welcome)
  - [What is PerfOwl?](#what-is-perfowl)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
  - [Installation](#installation)
    - [Build from Source](#build-from-source)
    - [Docker](#docker)
  - [CLI Usage](#cli-usage)
    - [Available Commands](#available-commands)
    - [Output Formats](#output-formats)
  - [MCP Server Setup](#mcp-server-setup)
    - [Available MCP Tools](#available-mcp-tools)
  - [Usage Examples](#usage-examples)
    - [Analyzing a Profile](#analyzing-a-profile)
    - [Comparing Profiles](#comparing-profiles)
    - [Working with AI Assistants](#working-with-ai-assistants)
  - [Contributing](#contributing)
  - [Reporting Issues](#reporting-issues)
  - [License](#license)

## Prerequisites

Before using PerfOwl, ensure you have:

- **Go >= 1.21** - Required for building from source. [Download here](https://go.dev/dl/)
- **Docker** (optional) - For containerized usage. [Installation guide](https://docs.docker.com/engine/install/)

> [!NOTE]
> PerfOwl analyzes Firefox Profiler exports. To capture a profile:
>
> 1. Open Firefox and navigate to [profiler.firefox.com](https://profiler.firefox.com)
> 2. Click "Enable Profiler" and reproduce your performance issue
> 3. Click "Capture Profile" and download the `.json.gz` file

## Quick Start

```bash
# Clone and build
git clone https://github.com/user/perfowl.git
cd perfowl
go build -o perfowl .

# Analyze a profile
./perfowl summary -p profile.json.gz
./perfowl bottlenecks -p profile.json.gz -o markdown
```

## Installation

### Build from Source

```bash
go build -o perfowl .
```

### Docker

```bash
# Build image
docker build -t perfowl .

# Run CLI commands
docker run --rm -v /path/to/profiles:/profiles:ro perfowl summary -p /profiles/profile.json.gz

# Run MCP server
docker run -i --rm -v /path/to/profiles:/profiles:ro perfowl mcp
```

## CLI Usage

### Available Commands

|Command|Description|
|---|---|
|`summary`|Get profile summary (duration, platform, threads, extensions)|
|`bottlenecks`|Detect performance bottlenecks with severity filtering|
|`extensions`|Analyze extension performance impact|
|`markers`|Extract markers filtered by type, category, or duration|
|`workers`|Analyze Web Worker performance and synchronization|
|`crypto`|Profile cryptographic operations and detect issues|
|`contention`|Detect thread contention (GC, IPC, locks)|
|`scaling`|Measure parallel scaling efficiency|
|`mcp`|Start the MCP server|

### Output Formats

PerfOwl supports multiple output formats:

```bash
# JSON (default)
./perfowl bottlenecks -p profile.json.gz -o json

# Markdown (great for reports)
./perfowl bottlenecks -p profile.json.gz -o markdown

# Plain text
./perfowl bottlenecks -p profile.json.gz -o text
```

## MCP Server Setup

PerfOwl includes an MCP (Model Context Protocol) server for integration with AI assistants like Claude.

**Add to Claude Code:**

```bash
claude mcp add perfowl -s user -- /path/to/perfowl mcp
```

**Verify installation:**

```bash
claude mcp list
```

**Start using with Claude:**

> "Analyze the Firefox profile at /path/to/profile.json.gz and find bottlenecks"

### Available MCP Tools

|Tool|Description|
|---|---|
|`get_summary`|Get profile summary (duration, platform, threads, extensions)|
|`get_bottlenecks`|Detect performance bottlenecks with severity filtering|
|`get_markers`|Extract markers filtered by type, category, or duration|
|`analyze_extension`|Analyze extension performance impact|
|`analyze_profile`|Comprehensive analysis (summary + bottlenecks + extensions)|
|`get_call_tree`|Find hot functions and hot paths|
|`get_category_breakdown`|Time spent per profiler category|
|`get_thread_analysis`|Analyze all threads with CPU time and wake patterns|
|`compare_profiles`|Compare two profiles for improvements/regressions|
|`analyze_workers`|Analyze Web Worker performance and synchronization|
|`analyze_crypto`|Profile cryptographic operations and detect issues|
|`analyze_contention`|Detect thread contention (GC, IPC, locks)|
|`analyze_scaling`|Measure parallel scaling efficiency|
|`compare_scaling`|Compare scaling between two profiles|

## Usage Examples

### Analyzing a Profile

```bash
# Get a quick summary
./perfowl summary -p profile.json.gz

# Find performance bottlenecks
./perfowl bottlenecks -p profile.json.gz -o markdown

# Analyze extension impact
./perfowl extensions -p profile.json.gz

# Extract specific markers
./perfowl markers -p profile.json.gz --type GCMajor --limit 10

# Analyze worker threads
./perfowl workers -p profile.json.gz

# Profile crypto operations
./perfowl crypto -p profile.json.gz

# Detect thread contention
./perfowl contention -p profile.json.gz

# Measure scaling efficiency
./perfowl scaling -p profile.json.gz
```

### Comparing Profiles

```bash
# Compare scaling between baseline and optimized versions
./perfowl scaling -p baseline.json.gz --compare optimized.json.gz
```

### Working with AI Assistants

With the MCP server running, you can ask Claude to:

- "Analyze this profile and summarize the main bottlenecks"
- "Compare these two profiles and tell me what improved"
- "Find any extensions that are causing performance issues"
- "Check if there's GC pressure affecting the main thread"

## Contributing

Contributions are welcome! Here's how to get started:

1. **Fork the Repository** - Create your own copy of the project
2. **Create a Feature Branch** - `git checkout -b feature/your-feature-name`
3. **Make Your Changes** - Follow Go conventions and run `go fmt`
4. **Test Thoroughly** - Ensure your changes work correctly
5. **Submit a Pull Request** - Describe your changes and their impact

## Reporting Issues

Found a bug or have a feature request?

1. **Check Existing Issues** - Avoid duplicates by searching first
2. **Provide Details** - Include environment info, steps to reproduce, and expected behavior
3. **Share Sample Profiles** - If possible, share a minimal profile that reproduces the issue

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
